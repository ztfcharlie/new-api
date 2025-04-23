package model

import (
	"errors"
	"fmt"
	"one-api/common"
	"one-api/lang"
	"strings"

	"github.com/bytedance/gopkg/util/gopool"
)

// root 获取所有token
func GetRootAllUserTokens(userId int, startIdx int, num int) ([]*Token, error) {
	var tokens []*Token
	var err error
	db := DB
	err = db.Order("id desc").Limit(num).Offset(startIdx).Find(&tokens).Error
	return tokens, err
}

// 根据id获取详情
func GetRootTokenByIds(id int) (*Token, error) {
	if id == 0 {
		return nil, errors.New(lang.T(nil, "token.error.id_or_userid_empty"))
	}
	token := Token{Id: id}
	var err error = nil
	err = DB.First(&token, "id = ?", id).Error
	return &token, err
}

// 搜索token
func SearchRootUserTokens(userId int, keyword string, token string) (tokens []*Token, err error) {
	if token != "" {
		token = strings.Trim(token, "sk-")
	}
	db := DB
	if userId != 0 {
		db = db.Where("user_id = ?", userId)
	}
	if keyword != "" {
		db = db.Where("name LIKE ?", "%"+keyword+"%")
	}
	if token != "" {
		db = db.Where(keyCol+" LIKE ?", "%"+token+"%")
	}
	err = db.Find(&tokens).Error
	return tokens, err
}

// 删除
func DeleteRootTokenById(id int) (err error) {
	// Why we need userId here? In case user want to delete other's token.
	if id == 0 {
		return errors.New(lang.T(nil, "token.error.id_or_userid_empty"))
	}
	token := Token{Id: id}
	err = DB.Where(token).First(&token).Error
	if err != nil {
		return err
	}
	return token.Delete()
}

// 更新
func (token *Token) UpdateRoot() (err error) {
	defer func() {
		if shouldUpdateRedis(true, err) {
			gopool.Go(func() {
				err := cacheSetToken(*token)
				if err != nil {
					common.SysError(fmt.Sprintf(lang.T(nil, "token.error.cache_token_update"), err.Error()))
				}
			})
		}
	}()
	err = DB.Model(token).Select("name", "user_id", "status", "expired_time", "remain_quota", "unlimited_quota",
		"model_limits_enabled", "model_limits", "allow_ips", "group").Updates(token).Error
	return err
}
