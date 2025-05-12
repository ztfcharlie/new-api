package controller

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"one-api/common"
	"one-api/lang"
	"one-api/model"
	"one-api/service"
	"one-api/setting"
	"strconv"
	"sync"
	"time"

	"github.com/Calcium-Ion/go-epay/epay"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/webhook"
)

const (
	PayTypeWxPay    = "wxpay"
	PayTypeAliPay   = "alipay"
	PayTypeStripe   = "stripepay"
	PayTypeCoinbase = "coinbasepay"
	PayTypePaypal   = "paypalpay"
)

type EpayRequest struct {
	Amount        int64  `json:"amount"`
	PaymentMethod string `json:"payment_method"`
	TopUpCode     string `json:"top_up_code"`
}

type AmountRequest struct {
	Amount    int64  `json:"amount"`
	TopUpCode string `json:"top_up_code"`
}

func GetEpayClient() *epay.Client {
	if setting.PayAddress == "" || setting.EpayId == "" || setting.EpayKey == "" {
		return nil
	}
	withUrl, err := epay.NewClient(&epay.Config{
		PartnerID: setting.EpayId,
		Key:       setting.EpayKey,
	}, setting.PayAddress)
	if err != nil {
		return nil
	}
	return withUrl
}

func getPayMoney(amount int64, group string) float64 {
	dAmount := decimal.NewFromInt(amount)

	if !common.DisplayInCurrencyEnabled {
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		dAmount = dAmount.Div(dQuotaPerUnit)
	}

	topupGroupRatio := common.GetTopupGroupRatio(group)
	if topupGroupRatio == 0 {
		topupGroupRatio = 1
	}

	dTopupGroupRatio := decimal.NewFromFloat(topupGroupRatio)
	dPrice := decimal.NewFromFloat(setting.Price)

	payMoney := dAmount.Mul(dPrice).Mul(dTopupGroupRatio)

	return payMoney.InexactFloat64()
}

func getMinTopup() int64 {
	minTopup := setting.MinTopUp
	if !common.DisplayInCurrencyEnabled {
		dMinTopup := decimal.NewFromInt(int64(minTopup))
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		minTopup = int(dMinTopup.Mul(dQuotaPerUnit).IntPart())
	}
	return int64(minTopup)
}

func RequestEpay(c *gin.Context) {
	var req EpayRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": lang.T(c, "topup.error.params")})
		return
	}
	if req.Amount < getMinTopup() {
		c.JSON(200, gin.H{"message": "error", "data": fmt.Sprintf(lang.T(c, "topup.error.min_amount"), getMinTopup())})
		return
	}

	id := c.GetInt("id")
	group, err := model.GetUserGroup(id, true)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": lang.T(c, "topup.error.get_group")})
		return
	}
	payMoney := getPayMoney(req.Amount, group)
	if payMoney < 0.01 {
		c.JSON(200, gin.H{"message": "error", "data": lang.T(c, "topup.error.amount_too_low")})
		return
	}

	callBackAddress := service.GetCallbackAddress()
	returnUrl, _ := url.Parse(setting.ServerAddress + "/log")
	notifyUrl, _ := url.Parse(callBackAddress + "/api/user/epay/notify")
	tradeNo := fmt.Sprintf("%s%d", common.GetRandomString(6), time.Now().Unix())
	tradeNo = fmt.Sprintf("USR%dNO%s", id, tradeNo)

	payType := PayTypeWxPay
	switch req.PaymentMethod {
	case "zfb":
		payType = PayTypeAliPay
	case "wx", "wxpay":
		payType = PayTypeWxPay
	case "stripe":
		payType = PayTypeStripe
		notifyUrl, _ = url.Parse(callBackAddress + "/api/user/stripe/notify")
	case "coinbase":
		payType = PayTypeCoinbase
		notifyUrl, _ = url.Parse(callBackAddress + "/api/user/coinbase/notify")
	case "paypal":
		payType = PayTypePaypal
		notifyUrl, _ = url.Parse(callBackAddress + "/api/user/paypal/notify")
	case "epay":
		notifyUrl, _ = url.Parse(callBackAddress + "/api/user/epay/notify")
	}

	var uri string
	var params map[string]string

	switch payType {
	case PayTypeStripe:
		client := service.GetStripeClient()
		if client == nil {
			c.JSON(200, gin.H{"message": "error", "data": lang.T(c, "topup.error.no_payment_config")})
			return
		}
		uri, params, err = client.Purchase(&service.PurchaseArgs{
			Type:           payType,
			ServiceTradeNo: tradeNo,
			Name:           fmt.Sprintf("Bruncloud Credit Top-up %d", req.Amount),
			Money:          strconv.FormatFloat(payMoney, 'f', 2, 64),
			Device:         service.PC,
			NotifyUrl:      notifyUrl,
			ReturnUrl:      returnUrl,
		})
		params["method"] = "GET"
	case PayTypeCoinbase:
		client := service.GetCoinbaseClient()
		if client == nil {
			c.JSON(200, gin.H{"message": "error", "data": lang.T(c, "topup.error.coinbase_not_found")})
			return
		}
		uri, params, err = client.Purchase(&service.PurchaseArgs{
			Type:           payType,
			ServiceTradeNo: tradeNo,
			Name:           fmt.Sprintf("Bruncloud Credit Top-up %d", req.Amount),
			Money:          strconv.FormatFloat(payMoney, 'f', 2, 64),
			Device:         service.PC,
			NotifyUrl:      notifyUrl,
			ReturnUrl:      returnUrl,
		})
		params["method"] = "GET"
	case PayTypePaypal:
		client := service.GetPaypalClient()
		if client == nil {
			c.JSON(200, gin.H{"message": "error", "data": lang.T(c, "topup.error.paypal_not_found")})
			return
		}
		uri, params, err = client.Purchase(&service.PurchaseArgs{
			Type:           payType,
			ServiceTradeNo: tradeNo,
			Name:           fmt.Sprintf("Bruncloud Credit Top-up %d", req.Amount),
			Money:          strconv.FormatFloat(payMoney, 'f', 2, 64),
			Device:         service.PC,
			NotifyUrl:      notifyUrl,
			ReturnUrl:      returnUrl,
		})
		if err != nil {
			translated := lang.T(c, "topup.error.paypal_create_faild")
			fmt.Printf("Debug - Translated result: '%s'\n", translated)
			c.JSON(200, gin.H{"message": "error", "data": translated})
			return
		}
		params["method"] = "GET"
	default:
		// 原易支付逻辑
		client := GetEpayClient()
		if client == nil {
			c.JSON(200, gin.H{"message": "error", "data": lang.T(c, "topup.error.no_payment_config")})
			return
		}
		//修改给前端显示的人民币价格
		RmbPriceRatio := 7.3 // 默认值
		if setting.RmbPrice > 0 {
			RmbPriceRatio = setting.RmbPrice
		}
		payMoney_RMB := payMoney * RmbPriceRatio
		uri, params, err = client.Purchase(&epay.PurchaseArgs{
			Type:           payType,
			ServiceTradeNo: tradeNo,
			Name:           fmt.Sprintf("Bruncloud Credit Top-up %d", req.Amount),
			Money:          strconv.FormatFloat(payMoney_RMB, 'f', 2, 64),
			Device:         epay.PC,
			NotifyUrl:      notifyUrl,
			ReturnUrl:      returnUrl,
		})
		params["method"] = "POST"
	}

	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": lang.T(c, "topup.error.payment_failed")})
		return
	}
	amount := req.Amount
	if !common.DisplayInCurrencyEnabled {
		dAmount := decimal.NewFromInt(int64(amount))
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		amount = dAmount.Div(dQuotaPerUnit).IntPart()
	}
	topUp := &model.TopUp{
		UserId:     id,
		Amount:     amount,
		Money:      payMoney,
		TradeNo:    tradeNo,
		CreateTime: time.Now().Unix(),
		Status:     "pending",
	}
	err = topUp.Insert()
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": lang.T(c, "topup.error.create_order")})
		return
	}
	c.JSON(200, gin.H{"message": "success", "data": params, "url": uri})
}

// tradeNo lock
var orderLocks sync.Map
var createLock sync.Mutex

// LockOrder 尝试对给定订单号加锁
func LockOrder(tradeNo string) {
	lock, ok := orderLocks.Load(tradeNo)
	if !ok {
		createLock.Lock()
		defer createLock.Unlock()
		lock, ok = orderLocks.Load(tradeNo)
		if !ok {
			lock = new(sync.Mutex)
			orderLocks.Store(tradeNo, lock)
		}
	}
	lock.(*sync.Mutex).Lock()
}

// UnlockOrder 释放给定订单号的锁
func UnlockOrder(tradeNo string) {
	lock, ok := orderLocks.Load(tradeNo)
	if ok {
		lock.(*sync.Mutex).Unlock()
	}
}

func EpayNotify(c *gin.Context) {
	params := lo.Reduce(lo.Keys(c.Request.URL.Query()), func(r map[string]string, t string, i int) map[string]string {
		r[t] = c.Request.URL.Query().Get(t)
		return r
	}, map[string]string{})
	client := GetEpayClient()
	if client == nil {
		log.Println(lang.T(c, "topup.log.epay_notify_failed"))
		_, err := c.Writer.Write([]byte("fail"))
		if err != nil {
			log.Println(lang.T(c, "topup.log.epay_write_failed"))
			return
		}
	}
	verifyInfo, err := client.Verify(params)
	if err == nil && verifyInfo.VerifyStatus {
		_, err := c.Writer.Write([]byte("success"))
		if err != nil {
			log.Println(lang.T(c, "topup.log.epay_write_failed"))
		}
	} else {
		_, err := c.Writer.Write([]byte("fail"))
		if err != nil {
			log.Println(lang.T(c, "topup.log.epay_write_failed"))
		}
		log.Println(lang.T(c, "topup.log.epay_verify_failed"))
		return
	}
	if verifyInfo.TradeStatus == epay.StatusTradeSuccess {
		info := VerifyInfo{
			ServiceTradeNo: verifyInfo.ServiceTradeNo,
		}
		err = handleOrderSuccess(info)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

	} else {
		log.Printf(lang.T(c, "topup.log.epay_callback_exception"), verifyInfo)

	}

}

type VerifyInfo struct {
	ServiceTradeNo string
}

func RequestAmount(c *gin.Context) {
	var req AmountRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": lang.T(c, "topup.error.params")})
		return
	}

	if req.Amount < getMinTopup() {
		c.JSON(200, gin.H{"message": "error", "data": fmt.Sprintf(lang.T(c, "topup.error.min_amount"), getMinTopup())})
		return
	}
	id := c.GetInt("id")
	group, err := model.GetUserGroup(id, true)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": lang.T(c, "topup.error.get_group")})
		return
	}
	payMoney := getPayMoney(req.Amount, group)
	if payMoney <= 0.01 {
		c.JSON(200, gin.H{"message": "error", "data": lang.T(c, "topup.error.amount_too_low")})
		return
	}
	c.JSON(200, gin.H{"message": "success", "data": strconv.FormatFloat(payMoney, 'f', 2, 64)})
}

// 处理订单通用方法
func handleOrderSuccess(verifyInfo VerifyInfo) error {
	log.Println(verifyInfo)
	LockOrder(verifyInfo.ServiceTradeNo)
	defer UnlockOrder(verifyInfo.ServiceTradeNo)
	topUp := model.GetTopUpByTradeNo(verifyInfo.ServiceTradeNo)
	if topUp == nil {
		return fmt.Errorf(lang.T(nil, "topup.log.epay_order_not_found"), verifyInfo)
	}
	if topUp.Status == "pending" {
		topUp.Status = "success"
		err := topUp.Update()
		if err != nil {
			return fmt.Errorf(lang.T(nil, "topup.log.epay_update_order_failed"), topUp)
		}
		//user, _ := model.GetUserById(topUp.UserId, false)
		//user.Quota += topUp.Amount * 500000
		quotaToAdd := int(float64(topUp.Amount) * common.QuotaPerUnit)
		err = model.IncreaseUserQuota(topUp.UserId, quotaToAdd, true)
		if err != nil {
			return fmt.Errorf(lang.T(nil, "topup.log.epay_update_user_failed"), topUp)
		}
		log.Printf(lang.T(nil, "topup.log.epay_update_success"), topUp)
		model.RecordLog(topUp.UserId, model.LogTypeTopup, fmt.Sprintf(lang.T(nil, "topup.record.success"), common.LogQuota(quotaToAdd), topUp.Money))

		currentUser, err := model.GetUserById(topUp.UserId, false)
		if err == nil {
			// 推荐人奖励
			if currentUser.InviterId > 0 {
				quotaValue := decimal.NewFromFloat(float64(quotaToAdd))
				QuotaForCount := decimal.NewFromFloat(float64(common.QuotaForCount))
				QuotaForCountFloat := QuotaForCount.Div(decimal.NewFromFloat(100))

				quotaForCountResult := quotaValue.Mul(QuotaForCountFloat).IntPart()
				if quotaForCountResult > 0 {
					err := model.UpdateUserAffQuota(nil, currentUser.InviterId, quotaForCountResult)
					if err != nil {
						log.Printf("奖励推荐人：%v，额度：%v，失败原因: %v\n", currentUser.InviterId, quotaForCountResult, err)
					} else {
						model.RecordLog(currentUser.InviterId, model.LogTypeInviterQuotaForCount, fmt.Sprintf(lang.T(nil, "user.log.inviter_quota_for_count"), quotaForCountResult))

					}
				}

			}
		}

		return nil
	}
	return fmt.Errorf(lang.T(nil, "topup.error.failed"), topUp)
}

func StripeNotify(c *gin.Context) {
	// Webhook 密钥
	stripeWebhookSecret := setting.StripeWebHookKey

	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// 验证签名
	stripeSignature := c.GetHeader("Stripe-Signature")
	event, err := webhook.ConstructEvent(body, stripeSignature, stripeWebhookSecret)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature"})
		return
	}

	// 处理事件
	switch event.Type {
	case "checkout.session.completed":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse payment intent"})
			return
		}
		// 处理支付成功的逻辑
		log.Printf("PaymentIntent failed: %s", paymentIntent.ID)
		trade_no := paymentIntent.Metadata["trade_no"]
		info := VerifyInfo{
			ServiceTradeNo: trade_no,
		}
		// 处理成功的订单
		err = handleOrderSuccess(info)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	default:
		log.Printf("Unhandled event type: %s", event.Type)
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})

}

// 添加 Coinbase 回调处理函数
func CoinbaseNotify(c *gin.Context) {
	// 读取 Coinbase Webhook 密钥
	coinbaseWebhookSecret := setting.CoinbaseWebHookKey

	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	//打印主题body
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// 验证签名
	signature := c.GetHeader("X-CC-Webhook-Signature")
	if err := verifyCoinbaseSignature(signature, body, coinbaseWebhookSecret); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature"})
		return
	}
	// 解析事件数据 - 修改为正确的结构
	var webhookData struct {
		Event struct {
			Type string `json:"type"`
			Data struct {
				Metadata map[string]string `json:"metadata"`
			} `json:"data"`
		} `json:"event"`
	}

	if err := json.Unmarshal(body, &webhookData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// 处理支付成功事件
	if webhookData.Event.Type == "charge:confirmed" {

		// 获取交易号
		tradeNo := webhookData.Event.Data.Metadata["trade_no"]
		if tradeNo == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing trade_no"})
			return
		}

		info := VerifyInfo{
			ServiceTradeNo: tradeNo,
		}
		err = handleOrderSuccess(info)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		log.Printf("CoinbaseNotify: Unhandled event type: %s", webhookData.Event.Type)
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// 添加 Coinbase 签名验证函数
func verifyCoinbaseSignature(signature string, payload []byte, secret string) error {
	// Coinbase 使用 HMAC-SHA256 进行签名验证
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// 比较签名
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}

// PayPal Webhook 回调处理函数
func PaypalNotify(c *gin.Context) {
	//log.Println("PaypalNotify: Starting PayPal Webhook handler")

	// 1. 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("PaypalNotify: Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	log.Println("PaypalNotify: Successfully read request body")
	//log.Printf("PaypalNotify: Raw webhook data: %s", string(body))

	// 2. 解析 JSON 数据
	var webhookData struct {
		EventType string `json:"event_type"`
		Resource  struct {
			ID            string `json:"id"`
			Status        string `json:"status"`
			PurchaseUnits []struct {
				ReferenceID string `json:"reference_id"`
				Payments    struct {
					Captures []struct {
						ID     string `json:"id"`
						Status string `json:"status"`
					} `json:"captures"`
				} `json:"payments"`
			} `json:"purchase_units"`
		} `json:"resource"`
	}

	if err := json.Unmarshal(body, &webhookData); err != nil {
		log.Printf("PaypalNotify: Failed to parse JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	//log.Printf("PaypalNotify: Parsed webhook data - event_type: %s, order_id: %s, status: %s", webhookData.EventType, webhookData.Resource.ID, webhookData.Resource.Status)

	// 3. 验证事件类型和支付状态
	if webhookData.EventType != "CHECKOUT.ORDER.APPROVED" {
		log.Printf("PaypalNotify: Unhandled event type: %s", webhookData.EventType)
		c.JSON(http.StatusOK, gin.H{"status": "success"}) // 返回成功，因为这是正常的通知
		return
	}

	if webhookData.Resource.Status != "APPROVED" {
		log.Printf("PaypalNotify: Order not completed, status: %s", webhookData.Resource.Status)
		c.JSON(http.StatusOK, gin.H{"status": "success"})
		return
	}

	log.Println("PaypalNotify: Order is completed, proceeding with processing")

	// 4. 查找交易号 - 直接使用 reference_id
	var tradeNo string

	// 检查 purchase_units 中的 referenceID
	if len(webhookData.Resource.PurchaseUnits) > 0 {
		tradeNo = webhookData.Resource.PurchaseUnits[0].ReferenceID
		log.Printf("PaypalNotify: Using reference_id as trade_no: %s", tradeNo)
	}

	// 5. 验证是否找到交易号
	if tradeNo == "" {
		log.Println("PaypalNotify: Could not find reference_id in webhook data")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing reference_id"})
		return
	}

	// 6. 处理订单
	log.Printf("PaypalNotify: Processing order with trade_no: %s", tradeNo)
	info := VerifyInfo{
		ServiceTradeNo: tradeNo,
	}

	err = handleOrderSuccess(info)
	if err != nil {
		log.Printf("PaypalNotify: Failed to handle order success: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Println("PaypalNotify: Order handled successfully")

	// 7. 返回成功响应
	log.Println("PaypalNotify: Sending success response")
	c.JSON(http.StatusOK, gin.H{"status": "success"})
	//log.Println("PaypalNotify: Webhook handler completed successfully")
}
