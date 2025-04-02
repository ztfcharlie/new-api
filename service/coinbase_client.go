package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/setting"
)

// CoinbaseClient 客户端
type CoinbaseClient struct {
}

// GetCoinbaseClient 获取 Coinbase 客户端实例
func GetCoinbaseClient() *CoinbaseClient {
	if setting.CoinbaseKey == "" {
		return nil
	}
	return &CoinbaseClient{}
}

type CoinbaseChargeRequest struct {
	Name         string            `json:"name"`
	Description  string           `json:"description"`
	PricingType  string           `json:"pricing_type"`
	LocalPrice   CoinbasePrice    `json:"local_price"`
	Metadata     map[string]string `json:"metadata"`
	RedirectURL  string           `json:"redirect_url"`
	CancelURL    string           `json:"cancel_url"`
}

type CoinbasePrice struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type CoinbaseChargeResponse struct {
	Data struct {
		HostedURL string `json:"hosted_url"`
	} `json:"data"`
}

// Purchase 创建支付会话
func (c *CoinbaseClient) Purchase(args *PurchaseArgs) (string, map[string]string, error) {
	apiURL := "https://api.commerce.coinbase.com/charges"
	
	// 构建支付请求
	chargeRequest := CoinbaseChargeRequest{
		Name:        args.Name,
		Description: "Payment for " + args.Name,
		PricingType: "fixed_price",
		LocalPrice: CoinbasePrice{
			Amount:   args.Money,
			Currency: "USD",
		},
		Metadata: map[string]string{
			"trade_no":   args.ServiceTradeNo,
			"order_name": args.Name,
		},
		RedirectURL: args.ReturnUrl.String(),
		CancelURL:   args.ReturnUrl.String(),
	}

	// 将请求转换为 JSON
	jsonData, err := json.Marshal(chargeRequest)
	if err != nil {
		return "", nil, fmt.Errorf("error marshaling charge request: %v", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", nil, fmt.Errorf("error creating request: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CC-Api-Key", setting.CoinbaseKey)
	req.Header.Set("X-CC-Version", "2018-03-22")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var chargeResponse CoinbaseChargeResponse
	if err := json.NewDecoder(resp.Body).Decode(&chargeResponse); err != nil {
		return "", nil, fmt.Errorf("error decoding response: %v", err)
	}

	// 构建请求参数
	requestParams := map[string]string{
		"pid":          "",
		"type":         args.Type,
		"out_trade_no": args.ServiceTradeNo,
		"notify_url":   args.NotifyUrl.String(),
		"name":         args.Name,
		"money":        args.Money,
		"device":       string(args.Device),
		"sign_type":    "MD5",
		"return_url":   args.ReturnUrl.String(),
		"sign":         "",
	}

	return chargeResponse.Data.HostedURL, requestParams, nil
}