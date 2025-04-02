package controller

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"one-api/common"
	"one-api/lang"
	"one-api/model"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LinuxdoUser struct {
	Id         int    `json:"id"`
	Username   string `json:"username"`
	Name       string `json:"name"`
	Active     bool   `json:"active"`
	TrustLevel int    `json:"trust_level"`
	Silenced   bool   `json:"silenced"`
}

func LinuxDoBind(c *gin.Context) {
	if !common.LinuxDOOAuthEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": lang.T(c, "linuxdo.error.oauth_disabled"),
		})
		return
	}

	code := c.Query("code")
	linuxdoUser, err := getLinuxdoUserInfoByCode(code, c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	user := model.User{
		LinuxDOId: strconv.Itoa(linuxdoUser.Id),
	}

	if model.IsLinuxDOIdAlreadyTaken(user.LinuxDOId) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": lang.T(c, "linuxdo.error.already_bound"),
		})
		return
	}

	session := sessions.Default(c)
	id := session.Get("id")
	user.Id = id.(int)

	err = user.FillUserById()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	user.LinuxDOId = strconv.Itoa(linuxdoUser.Id)
	err = user.Update(false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "bind",
	})
}

func getLinuxdoUserInfoByCode(code string, c *gin.Context) (*LinuxdoUser, error) {
	if code == "" {
		return nil, errors.New(lang.T(c, "linuxdo.error.invalid_code"))
	}

	// Get access token using Basic auth
	tokenEndpoint := "https://connect.linux.do/oauth2/token"
	credentials := common.LinuxDOClientId + ":" + common.LinuxDOClientSecret
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(credentials))

	// Get redirect URI from request
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	redirectURI := fmt.Sprintf("%s://%s/api/oauth/linuxdo", scheme, c.Request.Host)

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequest("POST", tokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", basicAuth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := http.Client{Timeout: 5 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.New(lang.T(c, "linuxdo.error.connect_failed"))
	}
	defer res.Body.Close()

	var tokenRes struct {
		AccessToken string `json:"access_token"`
		Message     string `json:"message"`
	}
	if err := json.NewDecoder(res.Body).Decode(&tokenRes); err != nil {
		return nil, err
	}

	if tokenRes.AccessToken == "" {
		return nil, fmt.Errorf(lang.T(c, "linuxdo.error.token_failed"), tokenRes.Message)
	}

	// Get user info
	userEndpoint := "https://connect.linux.do/api/user"
	req, err = http.NewRequest("GET", userEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tokenRes.AccessToken)
	req.Header.Set("Accept", "application/json")

	res2, err := client.Do(req)
	if err != nil {
		return nil, errors.New(lang.T(c, "linuxdo.error.get_user_failed"))
	}
	defer res2.Body.Close()

	var linuxdoUser LinuxdoUser
	if err := json.NewDecoder(res2.Body).Decode(&linuxdoUser); err != nil {
		return nil, err
	}

	if linuxdoUser.Id == 0 {
		return nil, errors.New(lang.T(c, "linuxdo.error.invalid_user"))
	}

	return &linuxdoUser, nil
}

func LinuxdoOAuth(c *gin.Context) {
	session := sessions.Default(c)

	errorCode := c.Query("error")
	if errorCode != "" {
		errorDescription := c.Query("error_description")
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": errorDescription,
		})
		return
	}

	state := c.Query("state")
	if state == "" || session.Get("oauth_state") == nil || state != session.Get("oauth_state").(string) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": lang.T(c, "linuxdo.error.state_invalid"),
		})
		return
	}

	username := session.Get("username")
	if username != nil {
		LinuxDoBind(c)
		return
	}

	if !common.LinuxDOOAuthEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": lang.T(c, "linuxdo.error.oauth_disabled"),
		})
		return
	}

	code := c.Query("code")
	linuxdoUser, err := getLinuxdoUserInfoByCode(code, c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	user := model.User{
		LinuxDOId: strconv.Itoa(linuxdoUser.Id),
	}

	// Check if user exists
	if model.IsLinuxDOIdAlreadyTaken(user.LinuxDOId) {
		err := user.FillUserByLinuxDOId()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
		if user.Id == 0 {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": lang.T(c, "linuxdo.error.user_deleted"),
			})
			return
		}
	} else {
		if common.RegisterEnabled {
			user.Username = "linuxdo_" + strconv.Itoa(model.GetMaxUserId()+1)
			user.DisplayName = linuxdoUser.Name
			user.Role = common.RoleCommonUser
			user.Status = common.UserStatusEnabled

			affCode := session.Get("aff")
			inviterId := 0
			if affCode != nil {
				inviterId, _ = model.GetUserIdByAffCode(affCode.(string))
			}

			if err := user.Insert(inviterId); err != nil {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": err.Error(),
				})
				return
			}
		} else {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": lang.T(c, "user.register.disabled"),
			})
			return
		}
	}

	if user.Status != common.UserStatusEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": lang.T(c, "linuxdo.error.user_banned"),
			"success": false,
		})
		return
	}

	setupLogin(&user, c)
}
