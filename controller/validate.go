package controller

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

// 定义一个全局翻译器T
var trans ut.Translator

// InitTrans 初始化翻译器
func initTrans(locale string) (err error) {
	// 修改gin框架中的Validator引擎属性，实现自定制
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 注册一个获取json tag的自定义方法
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})

		zhT := zh.New() // 中文翻译器
		enT := en.New() // 英文翻译器

		// 第一个参数是备用（fallback）的语言环境
		// 后面的参数是应该支持的语言环境（支持多个）
		// uni := ut.New(zhT, zhT) 也是可以的
		uni := ut.New(enT, zhT, enT)

		// locale 通常取决于 http 请求头的 'Accept-Language'
		var ok bool
		// 也可以使用 uni.FindTranslator(...) 传入多个locale进行查找
		trans, ok = uni.GetTranslator(locale)
		if !ok {
			return fmt.Errorf("uni.GetTranslator(%s) failed", locale)
		}

		// 注册翻译器
		switch locale {
		case "zh":
			err = zh_translations.RegisterDefaultTranslations(v, trans)
		case "en":
			err = en_translations.RegisterDefaultTranslations(v, trans)
		default:
			err = en_translations.RegisterDefaultTranslations(v, trans)
		}
		return
	}
	return
}

// RemoveTopStruct 去除字段名中的结构体名称标识
func removeTopStruct(fields map[string]string) map[string]string {
	res := map[string]string{}
	for field, err := range fields {
		res[field[strings.Index(field, ".")+1:]] = err
	}
	return res
}

// 获取表单错误信息
func getValidateError(lang string, err error) error {
	// 初始化翻译器
	if err := initTrans(lang); err != nil {
		return errors.New("初始化翻译器失败")
	}
	// 获取验证错误
	if vError, ok := err.(validator.ValidationErrors); ok {
		errMap := vError.Translate(trans)
		// 去除结构体名称前缀
		cleanErr := removeTopStruct(errMap)
		for _, v := range cleanErr {
			return errors.New(v)
		}
		return errors.New("参数错误")
	}
	return errors.New("请求参数错误")
}
