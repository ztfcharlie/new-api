package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/lang"
	"one-api/model"
	"one-api/setting"
	"one-api/setting/operation_setting"
	"one-api/setting/system_setting"
	"strings"

	"github.com/gin-gonic/gin"
)

func TestStatus(c *gin.Context) {
	err := model.PingDB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"message": lang.T(c, "misc.error.db_connection"),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": lang.T(c, "misc.status.running"),
	})
	return
}

func GetStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"version":                     common.Version,
			"start_time":                  common.StartTime,
			"email_verification":          common.EmailVerificationEnabled,
			"github_oauth":                common.GitHubOAuthEnabled,
			"github_client_id":            common.GitHubClientId,
			"linuxdo_oauth":               common.LinuxDOOAuthEnabled,
			"linuxdo_client_id":           common.LinuxDOClientId,
			"telegram_oauth":              common.TelegramOAuthEnabled,
			"telegram_bot_name":           common.TelegramBotName,
			"system_name":                 common.SystemName,
			"logo":                        common.Logo,
			"footer_html":                 common.Footer,
			"wechat_qrcode":               common.WeChatAccountQRCodeImageURL,
			"wechat_login":                common.WeChatAuthEnabled,
			"server_address":              setting.ServerAddress,
			"price":                       setting.Price,
			"min_topup":                   setting.MinTopUp,
			"turnstile_check":             common.TurnstileCheckEnabled,
			"turnstile_site_key":          common.TurnstileSiteKey,
			"top_up_link":                 common.TopUpLink,
			"docs_link":                   operation_setting.GetGeneralSetting().DocsLink,
			"quota_per_unit":              common.QuotaPerUnit,
			"display_in_currency":         common.DisplayInCurrencyEnabled,
			"enable_batch_update":         common.BatchUpdateEnabled,
			"enable_drawing":              common.DrawingEnabled,
			"enable_task":                 common.TaskEnabled,
			"enable_data_export":          common.DataExportEnabled,
			"data_export_default_time":    common.DataExportDefaultTime,
			"default_collapse_sidebar":    common.DefaultCollapseSidebar,
			"enable_online_topup":         setting.PayAddress != "" && setting.EpayId != "" && setting.EpayKey != "",
			"mj_notify_enabled":           setting.MjNotifyEnabled,
			"chats":                       setting.Chats,
			"demo_site_enabled":           operation_setting.DemoSiteEnabled,
			"self_use_mode_enabled":       operation_setting.SelfUseModeEnabled,
			"oidc_enabled":                system_setting.GetOIDCSettings().Enabled,
			"oidc_client_id":              system_setting.GetOIDCSettings().ClientId,
			"oidc_authorization_endpoint": system_setting.GetOIDCSettings().AuthorizationEndpoint,
		},
	})
	return
}

func GetNotice(c *gin.Context) {
	common.OptionMapRWMutex.RLock()
	defer common.OptionMapRWMutex.RUnlock()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    common.OptionMap["Notice"],
	})
	return
}

func GetAbout(c *gin.Context) {
	common.OptionMapRWMutex.RLock()
	defer common.OptionMapRWMutex.RUnlock()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    common.OptionMap["About"],
	})
	return
}

func GetMidjourney(c *gin.Context) {
	common.OptionMapRWMutex.RLock()
	defer common.OptionMapRWMutex.RUnlock()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    common.OptionMap["Midjourney"],
	})
	return
}

func GetHomePageContent(c *gin.Context) {
	common.OptionMapRWMutex.RLock()
	defer common.OptionMapRWMutex.RUnlock()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    common.OptionMap["HomePageContent"],
	})
	return
}

func SendEmailVerification(c *gin.Context) {
	email := c.Query("email")
	if err := common.Validate.Var(email, "required,email"); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": lang.T(c, "misc.error.invalid_params"),
		})
		return
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": lang.T(c, "misc.error.invalid_email"),
		})
		return
	}
	localPart := parts[0]
	domainPart := parts[1]
	if common.EmailDomainRestrictionEnabled {
		allowed := false
		for _, domain := range common.EmailDomainWhitelist {
			if domainPart == domain {
				allowed = true
				break
			}
		}
		if !allowed {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": lang.T(c, "misc.error.email_domain_restricted"),
			})
			return
		}
	}
	if common.EmailAliasRestrictionEnabled {
		containsSpecialSymbols := strings.Contains(localPart, "+") || strings.Contains(localPart, ".")
		if containsSpecialSymbols {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": lang.T(c, "misc.error.email_alias_restricted"),
			})
			return
		}
	}

	if model.IsEmailAlreadyTaken(email) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": lang.T(c, "misc.error.email_taken"),
		})
		return
	}
	code := common.GenerateVerificationCode(6)
	common.RegisterVerificationCodeWithKey(email, code, common.EmailVerificationPurpose)
	subject := fmt.Sprintf(lang.T(c, "misc.email.verification_subject"), common.SystemName)
	content := fmt.Sprintf(lang.T(c, "misc.email.verification_content"),
		common.SystemName, code, common.VerificationValidMinutes)
	err := common.SendEmail(subject, email, content)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func SendPasswordResetEmail(c *gin.Context) {
	email := c.Query("email")
	if err := common.Validate.Var(email, "required,email"); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": lang.T(c, "misc.error.invalid_params"),
		})
		return
	}
	if !model.IsEmailAlreadyTaken(email) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": lang.T(c, "misc.error.email_not_registered"),
		})
		return
	}
	code := common.GenerateVerificationCode(0)
	common.RegisterVerificationCodeWithKey(email, code, common.PasswordResetPurpose)
	link := fmt.Sprintf("%s/user/reset?email=%s&token=%s", setting.ServerAddress, email, code)
	subject := fmt.Sprintf(lang.T(c, "misc.email.reset_subject"), common.SystemName)
	content := fmt.Sprintf(lang.T(c, "misc.email.reset_content"),
		common.SystemName, link, link, common.VerificationValidMinutes)
	err := common.SendEmail(subject, email, content)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

type PasswordResetRequest struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

func ResetPassword(c *gin.Context) {
	var req PasswordResetRequest
	err := json.NewDecoder(c.Request.Body).Decode(&req)
	if req.Email == "" || req.Token == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": lang.T(c, "misc.error.invalid_params"),
		})
		return
	}
	if !common.VerifyCodeWithKey(req.Email, req.Token, common.PasswordResetPurpose) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": lang.T(c, "misc.error.reset_link_invalid"),
		})
		return
	}
	password := common.GenerateVerificationCode(12)
	err = model.ResetUserPasswordByEmail(req.Email, password)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	common.DeleteKey(req.Email, common.PasswordResetPurpose)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    password,
	})
	return
}
