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
			Name:           fmt.Sprintf("TUC%d", req.Amount),
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
			Name:           fmt.Sprintf("TUC%d", req.Amount),
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
			Name:           fmt.Sprintf("TUC%d", req.Amount),
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
		uri, params, err = client.Purchase(&epay.PurchaseArgs{
			Type:           payType,
			ServiceTradeNo: tradeNo,
			Name:           fmt.Sprintf("TUC%d", req.Amount),
			Money:          strconv.FormatFloat(payMoney, 'f', 2, 64),
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
		log.Println(verifyInfo)
		LockOrder(verifyInfo.ServiceTradeNo)
		defer UnlockOrder(verifyInfo.ServiceTradeNo)
		topUp := model.GetTopUpByTradeNo(verifyInfo.ServiceTradeNo)
		if topUp == nil {
			log.Printf(lang.T(c, "topup.log.epay_order_not_found"), verifyInfo)
			return
		}
		if topUp.Status == "pending" {
			topUp.Status = "success"
			err := topUp.Update()
			if err != nil {
				log.Printf(lang.T(c, "topup.log.epay_update_order_failed"), topUp)
				return
			}
			//user, _ := model.GetUserById(topUp.UserId, false)
			//user.Quota += topUp.Amount * 500000
			// 另一种修复方式
			quotaToAdd := int(float64(topUp.Amount) * common.QuotaPerUnit)
			err = model.IncreaseUserQuota(topUp.UserId, quotaToAdd, true)
			if err != nil {
				log.Printf(lang.T(c, "topup.log.epay_update_user_failed"), topUp)
				return
			}
			log.Printf(lang.T(c, "topup.log.epay_update_success"), topUp)
			model.RecordLog(topUp.UserId, model.LogTypeTopup, fmt.Sprintf(lang.T(c, "topup.record.success"), common.LogQuota(quotaToAdd), topUp.Money))
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
		return nil
	}
	return fmt.Errorf(lang.T(nil, "topup.error.failed"), topUp)
}

// 添加 Coinbase 回调处理函数
func CoinbaseNotify(c *gin.Context) {
	// 读取 Coinbase Webhook 密钥
	coinbaseWebhookSecret := setting.CoinbaseWebHookKey

	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
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

	// 解析事件数据
	var event struct {
		Type string `json:"type"`
		Data struct {
			Metadata struct {
				TradeNo string `json:"trade_no"`
			} `json:"metadata"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// 处理支付成功事件
	if event.Type == "charge:confirmed" {
		info := VerifyInfo{
			ServiceTradeNo: event.Data.Metadata.TradeNo,
		}
		err = handleOrderSuccess(info)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
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

// 添加 PayPal 回调处理函数
// 添加 PayPal IPN 回调处理函数
func PaypalNotify(c *gin.Context) {
	// 1. 首先读取原始 POST 数据
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf(lang.T(c, "topup.log.paypal_read_body_failed"), err)
		c.String(http.StatusBadRequest, "Failed")
		return
	}

	// 2. 记录接收到的 IPN 数据
	log.Printf(lang.T(c, "topup.log.paypal_received_notification"), string(body))

	// 3. 解析 POST 数据
	err = c.Request.ParseForm()
	if err != nil {
		log.Printf(lang.T(c, "topup.log.paypal_parse_form_failed"), err)
		c.String(http.StatusBadRequest, "Failed")
		return
	}

	// 4. 获取关键参数
	paymentStatus := c.Request.PostForm.Get("payment_status")
	txnId := c.Request.PostForm.Get("txn_id")
	customData := c.Request.PostForm.Get("custom") // 这里会收到我们之前传入的 ServiceTradeNo

	// 5. 验证支付状态
	if paymentStatus != "Completed" {
		log.Printf("PayPal IPN: 支付未完成，状态: %s", paymentStatus)
		c.String(http.StatusOK, "OK") // 仍然返回成功，因为这是正常的通知
		return
	}

	// 6. 处理订单
	info := VerifyInfo{
		ServiceTradeNo: customData,
	}

	err = handleOrderSuccess(info)
	if err != nil {
		log.Printf("PayPal IPN: 处理订单失败: %v", err)
		c.String(http.StatusInternalServerError, "Failed")
		return
	}

	// 7. 记录成功日志
	log.Printf("PayPal IPN: 订单处理成功，交易号: %s, 订单号: %s", txnId, customData)

	// 8. 返回成功响应
	c.String(http.StatusOK, "OK")
}
