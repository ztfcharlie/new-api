package service

import (
    "bytes"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "one-api/setting"
    "strconv"
    "strings"
    "time"
)

const (
    PayPalSandboxAPI = "https://api.sandbox.paypal.com"
    PayPalLiveAPI    = "https://api.paypal.com"
)

// PayPalErrorResponse PayPal API 错误响应结构
type PayPalErrorResponse struct {
    Name        string `json:"name"`
    Message     string `json:"message"`
    DebugID     string `json:"debug_id"`
    Details     []struct {
        Field       string `json:"field"`
        Value       string `json:"value"`
        Location    string `json:"location"`
        Issue       string `json:"issue"`
        Description string `json:"description"`
    } `json:"details"`
}

// PayPalClient PayPal 客户端结构
type PayPalClient struct {
    clientID  string // PayPal Client ID
    secretKey string // PayPal Secret Key
    isSandbox bool   // 是否为沙盒环境
}

// GetPaypalClient 获取 PayPal 客户端实例
func GetPaypalClient() *PayPalClient {
    if setting.PaypalKey == "" || setting.PaypalWebHookKey == "" {
        return nil
    }
    return &PayPalClient{
        clientID:  setting.PaypalKey,
        secretKey: setting.PaypalWebHookKey,
        isSandbox: true, // 设置为 true 使用沙盒环境，生产环境时改为 false
    }
}

// 获取当前环境的 API 基础 URL
func (c *PayPalClient) getBaseURL() string {
    if c.isSandbox {
        return PayPalSandboxAPI
    }
    return PayPalLiveAPI
}

// 获取访问令牌
func (c *PayPalClient) getAccessToken() (string, error) {
    fmt.Println("Getting PayPal access token from", c.getBaseURL())

    // 构建认证信息
    auth := base64.StdEncoding.EncodeToString([]byte(c.clientID + ":" + c.secretKey))

    // 创建请求
    tokenURL := c.getBaseURL() + "/v1/oauth2/token"
    req, err := http.NewRequest("POST", tokenURL, strings.NewReader("grant_type=client_credentials"))
    if err != nil {
        return "", fmt.Errorf("failed to create token request: %v", err)
    }

    // 设置请求头
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Set("Authorization", "Basic "+auth)

    // 发送请求
    client := &http.Client{
        Timeout: 10 * time.Second,
    }
    resp, err := client.Do(req)
    if err != nil {
        return "", fmt.Errorf("failed to get access token: %v", err)
    }
    defer resp.Body.Close()

    // 解析响应
    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", fmt.Errorf("failed to parse token response: %v", err)
    }

    // 检查是否有错误
    if result["error"] != nil {
        return "", fmt.Errorf("PayPal error: %v", result["error_description"])
    }

    // 获取访问令牌
    accessToken, ok := result["access_token"].(string)
    if !ok {
        return "", fmt.Errorf("access token not found in response")
    }

    return accessToken, nil
}

// Purchase 创建 PayPal 支付订单
// Purchase 创建 PayPal 支付订单
func (c *PayPalClient) Purchase(args *PurchaseArgs) (string, map[string]string, error) {
    // 基本参数验证
    if args == nil {
        return "", nil, fmt.Errorf("purchase arguments cannot be nil")
    }

    // 将 Money 转换为 float64 进行验证
    amount, err := strconv.ParseFloat(args.Money, 64)
    if err != nil {
        return "", nil, fmt.Errorf("invalid amount format: %v", err)
    }
    if amount <= 0 {
        return "", nil, fmt.Errorf("invalid amount: must be greater than 0")
    }

    if args.ReturnUrl == nil {
        return "", nil, fmt.Errorf("return URL is required")
    }
    if !strings.HasPrefix(args.ReturnUrl.String(), "http://") && !strings.HasPrefix(args.ReturnUrl.String(), "https://") {
        return "", nil, fmt.Errorf("invalid return URL format: must start with http:// or https://")
    }

    // 获取访问令牌
    accessToken, err := c.getAccessToken()
    if err != nil {
        return "", nil, fmt.Errorf("failed to get PayPal access token: %v", err)
    }

    // 确保金额格式正确（保留两位小数）
    formattedAmount := fmt.Sprintf("%.2f", amount)

    // 构建 PayPal v2 API 支付请求
    paymentData := map[string]interface{}{
        "intent": "CAPTURE", // v2 API 使用 CAPTURE 而不是 sale
        "purchase_units": []map[string]interface{}{
            {
                "reference_id": args.ServiceTradeNo,
                "description": args.Name,
                "amount": map[string]interface{}{
                    "currency_code": "USD",
                    "value": formattedAmount,
                    "breakdown": map[string]interface{}{
                        "item_total": map[string]interface{}{
                            "currency_code": "USD",
                            "value": formattedAmount,
                        },
                    },
                },
                "items": []map[string]interface{}{
                    {
                        "name": args.Name,
                        "quantity": "1",
                        "unit_amount": map[string]interface{}{
                            "currency_code": "USD",
                            "value": formattedAmount,
                        },
                        "sku": args.ServiceTradeNo,
                    },
                },
            },
        },
        "application_context": map[string]interface{}{
            "return_url": fmt.Sprintf("%s?success=true&out_trade_no=%s", args.ReturnUrl.String(), args.ServiceTradeNo),
            "cancel_url": fmt.Sprintf("%s?success=false&out_trade_no=%s", args.ReturnUrl.String(), args.ServiceTradeNo),
            "brand_name": "BurnCloud",
            "shipping_preference": "NO_SHIPPING",
            "user_action": "PAY_NOW",
            "landing_page": "LOGIN",
        },
    }

    // 打印请求数据以便调试
    requestBody, _ := json.MarshalIndent(paymentData, "", "  ")
    fmt.Printf("PayPal Request Body: %s\n", string(requestBody))

    // 发送请求到 PayPal API
    paymentURL := c.getBaseURL() + "/v2/checkout/orders"
    jsonData, err := json.Marshal(paymentData)
    if err != nil {
        return "", nil, fmt.Errorf("failed to marshal payment data: %v", err)
    }

    req, err := http.NewRequest("POST", paymentURL, bytes.NewBuffer(jsonData))
    if err != nil {
        return "", nil, fmt.Errorf("failed to create request: %v", err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+accessToken)
    req.Header.Set("Prefer", "return=representation")

    client := &http.Client{
        Timeout: 30 * time.Second,
    }
    resp, err := client.Do(req)
    if err != nil {
        return "", nil, fmt.Errorf("failed to send request: %v", err)
    }
    defer resp.Body.Close()

    // 读取完整的响应体
    responseBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", nil, fmt.Errorf("failed to read response body: %v", err)
    }

    // 检查响应状态码
    if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
        var errorResponse PayPalErrorResponse
        if err := json.Unmarshal(responseBody, &errorResponse); err != nil {
            return "", nil, fmt.Errorf("PayPal API error: status code %d, body: %s", resp.StatusCode, string(responseBody))
        }
        return "", nil, fmt.Errorf("PayPal API error: %s - %s", errorResponse.Name, errorResponse.Message)
    }

    // 解析成功响应
    var result map[string]interface{}
    if err := json.Unmarshal(responseBody, &result); err != nil {
        return "", nil, fmt.Errorf("failed to parse response: %v", err)
    }

    // 打印响应以便调试
    fmt.Printf("PayPal API Response: %+v\n", result)

    // 从响应中获取支付URL
    var approvalURL string
    if links, ok := result["links"].([]interface{}); ok {
        for _, link := range links {
            if linkMap, ok := link.(map[string]interface{}); ok {
                if linkMap["rel"] == "approve" {
                    approvalURL = linkMap["href"].(string)
                    break
                }
            }
        }
    }

    if approvalURL == "" {
        return "", nil, fmt.Errorf("payment link not found in response: %+v", result)
    }

    return approvalURL, map[string]string{"method": "GET"}, nil
}