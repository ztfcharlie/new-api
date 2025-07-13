package controller

import (
	crypto_rand "crypto/rand" // 添加这一行
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/lang"
	"one-api/model"
	"strconv"

	// 添加这一行
	"github.com/gin-gonic/gin"
)

func GetAllTokens(c *gin.Context) {
	userId := c.GetInt("id")
	p, _ := strconv.Atoi(c.Query("p"))
	size, _ := strconv.Atoi(c.Query("size"))
	if p < 1 {
		p = 1
	}
	if size <= 0 {
		size = common.ItemsPerPage
	} else if size > 100 {
		size = 100
	}
	var tokens []*model.Token
	var err error
	if isRootUser(c) {
		tokens, err = model.GetRootAllUserTokens(0, (p-1)*size, size)
	} else {
		tokens, err = model.GetAllUserTokens(userId, (p-1)*size, size)
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	// Get total count for pagination
	total, _ := model.CountUserTokens(userId)

	// 添加token所属的用户字段
	userTokens, err := addTokenUsername(tokens)
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
		"data": gin.H{
			"items":     userTokens,
			"total":     total,
			"page":      p,
			"page_size": size,
		},
	})
	return
}

func SearchTokens(c *gin.Context) {
	userId := c.GetInt("id")
	keyword := c.Query("keyword")
	token := c.Query("token")
	username := c.Query("username")
	var tokens []*model.Token
	var err error
	if true || isRootUser(c) {
		if username != "" {
			searchUserId := -1
			searchUser, err := model.GetUserByUsername(username)
			if err == nil {
				searchUserId = searchUser.Id
			}
			tokens, err = model.SearchRootUserTokens(searchUserId, keyword, token)
		} else {
			tokens, err = model.SearchRootUserTokens(0, keyword, token)
		}
	} else {
		tokens, err = model.SearchUserTokens(userId, keyword, token)
	}

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	// 添加token所属的用户字段
	userTokens, err := addTokenUsername(tokens)
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
		"data":    userTokens,
	})
	return
}

func GetToken(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	userId := c.GetInt("id")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	var token *model.Token
	if isRootUser(c) {
		token, err = model.GetRootTokenByIds(id)
	} else {
		token, err = model.GetTokenByIds(id, userId)
	}
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
		"data":    token,
	})
	return
}

func GetTokenStatus(c *gin.Context) {
	tokenId := c.GetInt("token_id")
	userId := c.GetInt("id")
	var token *model.Token
	var err error
	if isRootUser(c) {
		token, err = model.GetRootTokenByIds(tokenId)
	} else {
		token, err = model.GetTokenByIds(tokenId, userId)
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	expiredAt := token.ExpiredTime
	if expiredAt == -1 {
		expiredAt = 0
	}
	c.JSON(http.StatusOK, gin.H{
		"object":          "credit_summary",
		"total_granted":   token.RemainQuota,
		"total_used":      0, // not supported currently
		"total_available": token.RemainQuota,
		"expires_at":      expiredAt * 1000,
	})
}

func AddToken(c *gin.Context) {
	token := model.Token{}
	err := c.ShouldBindJSON(&token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 处理名称为空的情况
	if token.Name == "" {
		// 直接在函数内实现生成6位随机字符串的逻辑
		const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		randomStr := make([]byte, 6)

		// 使用crypto/rand包生成安全的随机数
		randomData := make([]byte, 6)
		_, err := crypto_rand.Read(randomData)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Failed to generate random string: " + err.Error(),
			})
			return
		}

		// 使用随机数据从字符集中选择字符
		for i := range randomData {
			randomStr[i] = charset[randomData[i]%byte(len(charset))]
		}

		// 组合生成新的token名称
		token.Name = fmt.Sprintf("default_%s", string(randomStr))
	}

	if len(token.Name) > 30 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": lang.T(c, "token.error.name_too_long"),
		})
		return
	}
	key, err := common.GenerateKey()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": lang.T(c, "token.error.generate_failed"),
		})
		common.SysError(fmt.Sprintf(lang.T(c, "token.log.generate_failed"), err.Error()))
		return
	}
	cleanToken := model.Token{
		UserId:             c.GetInt("id"),
		Name:               token.Name,
		Key:                key,
		CreatedTime:        common.GetTimestamp(),
		AccessedTime:       common.GetTimestamp(),
		ExpiredTime:        token.ExpiredTime,
		RemainQuota:        token.RemainQuota,
		UnlimitedQuota:     token.UnlimitedQuota,
		ModelLimitsEnabled: token.ModelLimitsEnabled,
		ModelLimits:        token.ModelLimits,
		AllowIps:           token.AllowIps,
		Group:              token.Group,
	}
	if isRootUser(c) {
		cleanToken.UserId = token.UserId
	} else {
		cleanToken.UserId = c.GetInt("id")
	}
	err = cleanToken.Insert()
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

func DeleteToken(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userId := c.GetInt("id")
	var err error
	if isRootUser(c) {
		err = model.DeleteRootTokenById(id)
	} else {
		err = model.DeleteTokenById(id, userId)
	}
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

func UpdateToken(c *gin.Context) {
	userId := c.GetInt("id")
	statusOnly := c.Query("status_only")
	token := model.Token{}
	err := c.ShouldBindJSON(&token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if len(token.Name) > 30 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": lang.T(c, "token.error.name_too_long"),
		})
		return
	}
	var cleanToken *model.Token
	if isRootUser(c) {
		cleanToken, err = model.GetRootTokenByIds(token.Id)
	} else {
		cleanToken, err = model.GetTokenByIds(token.Id, userId)
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if token.Status == common.TokenStatusEnabled {
		if cleanToken.Status == common.TokenStatusExpired && cleanToken.ExpiredTime <= common.GetTimestamp() && cleanToken.ExpiredTime != -1 {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": lang.T(c, "token.error.expired"),
			})
			return
		}
		if cleanToken.Status == common.TokenStatusExhausted && cleanToken.RemainQuota <= 0 && !cleanToken.UnlimitedQuota {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": lang.T(c, "token.error.quota_exhausted"),
			})
			return
		}
	}
	if statusOnly != "" {
		cleanToken.Status = token.Status
	} else {
		// If you add more fields, please also update token.Update()
		cleanToken.Name = token.Name
		cleanToken.ExpiredTime = token.ExpiredTime
		cleanToken.RemainQuota = token.RemainQuota
		cleanToken.UnlimitedQuota = token.UnlimitedQuota
		cleanToken.ModelLimitsEnabled = token.ModelLimitsEnabled
		cleanToken.ModelLimits = token.ModelLimits
		cleanToken.AllowIps = token.AllowIps
		cleanToken.Group = token.Group
	}
	if isRootUser(c) {
		cleanToken.UserId = token.UserId
		err = cleanToken.UpdateRoot()
	} else {
		err = cleanToken.Update()
	}
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
		"data":    cleanToken,
	})
	return
}

type TokenBatch struct {
	Ids []int `json:"ids"`
}

func DeleteTokenBatch(c *gin.Context) {
	tokenBatch := TokenBatch{}
	if err := c.ShouldBindJSON(&tokenBatch); err != nil || len(tokenBatch.Ids) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "参数错误",
		})
		return
	}
	userId := c.GetInt("id")
	count, err := model.BatchDeleteTokens(tokenBatch.Ids, userId)
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
		"data":    count,
	})
}

type UserToken struct {
	model.Token
	Username string `json:"username"`
}

// 添加token所属的用户字段
func addTokenUsername(tokens []*model.Token) ([]*UserToken, error) {
	var userTokens []*UserToken
	if len(tokens) > 0 {
		var userIds []int
		for _, token := range tokens {
			fmt.Println(token.UserId)
			userIds = append(userIds, token.UserId)
		}
		users, err := model.GetUsersByIDs(userIds)
		userMap := make(map[int]string)
		for _, user := range users {
			userMap[user.Id] = user.Username
		}
		if err != nil {
			return nil, err
		}
		for _, v := range tokens {
			item := UserToken{
				Token: *v,
			}
			item.Username = userMap[v.UserId]
			userTokens = append(userTokens, &item)
		}
	}
	return userTokens, nil
}
