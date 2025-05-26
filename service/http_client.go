package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"one-api/common"
	"time"

	"golang.org/x/net/proxy"
)

var httpClient *http.Client
var impatientHTTPClient *http.Client

// ContentLengthFixer 是一个http.RoundTripper包装器，用于处理Content-Length不匹配问题
type ContentLengthFixer struct {
	Transport http.RoundTripper
}

// RoundTrip 实现http.RoundTripper接口
func (f *ContentLengthFixer) RoundTrip(req *http.Request) (*http.Response, error) {
	// 使用底层Transport执行请求
	resp, err := f.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// 如果响应有Content-Length头，但我们不确定它是否准确，
	// 我们可以通过读取整个响应体并重新设置它来解决这个问题
	if resp.Header.Get("Content-Length") != "" && resp.ContentLength > 0 && !resp.Close {
		// 读取整个响应体
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		// 重置响应体，使用正确的长度
		resp.Body = io.NopCloser(bytes.NewReader(body))
		resp.ContentLength = int64(len(body))
		resp.Header.Set("Content-Length", string(len(body)))

		// 确保不使用分块传输，因为我们现在有了准确的Content-Length
		resp.TransferEncoding = nil
	}

	return resp, nil
}

// 在发送非流式请求的地方使用修复后的HTTP客户端
// GetHTTPClientWithFixedContentLength 返回一个配置了ContentLengthFixer的HTTP客户端
func GetHTTPClientWithFixedContentLength() *http.Client {
	// 获取默认的HTTP客户端或您现有的客户端
	client := &http.Client{}

	// 应用我们的ContentLengthFixer
	if client.Transport == nil {
		client.Transport = &ContentLengthFixer{Transport: http.DefaultTransport}
	} else {
		client.Transport = &ContentLengthFixer{Transport: client.Transport}
	}

	return client
}

func init() {
	if common.RelayTimeout == 0 {
		httpClient = &http.Client{}
	} else {
		httpClient = &http.Client{
			Timeout: time.Duration(common.RelayTimeout) * time.Second,
		}
	}

	impatientHTTPClient = &http.Client{
		Timeout: 5 * time.Second,
	}
}

func GetHttpClient() *http.Client {
	return httpClient
}

func GetImpatientHttpClient() *http.Client {
	return impatientHTTPClient
}

// NewProxyHttpClient 创建支持代理的 HTTP 客户端
func NewProxyHttpClient(proxyURL string) (*http.Client, error) {
	if proxyURL == "" {
		return http.DefaultClient, nil
	}

	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	switch parsedURL.Scheme {
	case "http", "https":
		return &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(parsedURL),
			},
		}, nil

	case "socks5", "socks5h":
		// 获取认证信息
		var auth *proxy.Auth
		if parsedURL.User != nil {
			auth = &proxy.Auth{
				User:     parsedURL.User.Username(),
				Password: "",
			}
			if password, ok := parsedURL.User.Password(); ok {
				auth.Password = password
			}
		}

		// 创建 SOCKS5 代理拨号器
		// proxy.SOCKS5 使用 tcp 参数，所有 TCP 连接包括 DNS 查询都将通过代理进行。行为与 socks5h 相同
		dialer, err := proxy.SOCKS5("tcp", parsedURL.Host, auth, proxy.Direct)
		if err != nil {
			return nil, err
		}

		return &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					return dialer.Dial(network, addr)
				},
			},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported proxy scheme: %s", parsedURL.Scheme)
	}
}
