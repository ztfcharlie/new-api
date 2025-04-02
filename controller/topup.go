package controller

import (
	"fmt"
	"log"
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
	payType := "wxpay"
	if req.PaymentMethod == "zfb" {
		payType = "alipay"
	}
	if req.PaymentMethod == "wx" {
		req.PaymentMethod = "wxpay"
		payType = "wxpay"
	}
	callBackAddress := service.GetCallbackAddress()
	returnUrl, _ := url.Parse(setting.ServerAddress + "/log")
	notifyUrl, _ := url.Parse(callBackAddress + "/api/user/epay/notify")
	tradeNo := fmt.Sprintf("%s%d", common.GetRandomString(6), time.Now().Unix())
	tradeNo = fmt.Sprintf("USR%dNO%s", id, tradeNo)
	client := GetEpayClient()
	if client == nil {
		c.JSON(200, gin.H{"message": "error", "data": lang.T(c, "topup.error.no_payment_config")})
		return
	}
	uri, params, err := client.Purchase(&epay.PurchaseArgs{
		Type:           payType,
		ServiceTradeNo: tradeNo,
		Name:           fmt.Sprintf("TUC%d", req.Amount),
		Money:          strconv.FormatFloat(payMoney, 'f', 2, 64),
		Device:         epay.PC,
		NotifyUrl:      notifyUrl,
		ReturnUrl:      returnUrl,
	})
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
			dAmount := decimal.NewFromInt(int64(topUp.Amount))
			dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
			quotaToAdd := int(dAmount.Mul(dQuotaPerUnit).IntPart())
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
