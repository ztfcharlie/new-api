package service

import (
	"fmt"
	"math"
	"net/url"
	"one-api/setting"
	"strconv"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
)

// Stripe客户端
type StripeClient struct {
}
type DeviceType string

var (
	PC     DeviceType = "pc"     // PC PC端
	MOBILE DeviceType = "mobile" // MOBILE 移动端
)

type PurchaseArgs struct {
	// 支付类型
	Type string
	// 商家订单号
	ServiceTradeNo string
	// 商品名称
	Name string
	// 金额
	Money string
	// 设备类型
	Device    DeviceType
	NotifyUrl *url.URL
	ReturnUrl *url.URL
}

func GetStripeClient() *StripeClient {
	fmt.Println("setting.PayAddress", setting.StripeKey, setting.StripeWebHookKey)
	if setting.StripeKey == "" || setting.StripeWebHookKey == "" {
		return nil
	}
	return &StripeClient{}
}
func (stripeClient *StripeClient) Purchase(args *PurchaseArgs) (string, map[string]string, error) {

	// 设置 Stripe 密钥
	stripe.Key = setting.StripeKey

	// 创建 LineItem（商品信息）
	unitAmount, _ := strconv.ParseFloat(args.Money, 64)

	lineItem := &stripe.CheckoutSessionLineItemParams{
		Quantity: stripe.Int64(1),
		PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
			Currency: stripe.String("usd"), // 货币类型
			ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
				Name: stripe.String(args.Name), // 商品名称
			},
			UnitAmount: stripe.Int64(int64(math.Round(unitAmount * 100))), // 金额（以最小货币单位表示，例如1000表示10美元）
		},
	}

	// 将商品添加到 LineItems
	lineItems := []*stripe.CheckoutSessionLineItemParams{lineItem}

	// 创建 Checkout Session 参数
	params := &stripe.CheckoutSessionParams{
		SuccessURL:         stripe.String(args.ReturnUrl.String()), // 支付成功后的跳转链接
		CancelURL:          stripe.String(args.ReturnUrl.String()), // 支付取消后的跳转链接
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),   // 支付方式
		Mode:               stripe.String("payment"),               // 支付模式
		LineItems:          lineItems,                              // 商品列表
	}

	// 可选：添加元数据（例如订单ID）
	params.Metadata = map[string]string{
		"trade_no":   args.ServiceTradeNo,
		"order_name": args.Name,
	}

	// 创建 Checkout Session
	s, err := session.New(params)
	if err != nil {
		return "", nil, fmt.Errorf("Error creating Checkout Session: %v", err)
	}
	// 输出支付链接
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
	return s.URL, requestParams, nil
}
