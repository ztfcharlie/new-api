package lang

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/gin-gonic/gin"
)

const (
	// ContextKeyLang 在gin上下文中存储语言设置的键
	ContextKeyLang = "current_language"
	// DefaultLang 默认语言
	DefaultLang = "en"
)

var (
	translations = make(map[string]map[string]string)
	mutex        sync.RWMutex
)

// LoadTranslations 加载指定语言的翻译文件
func LoadTranslations(lang string) error {
	mutex.Lock()
	defer mutex.Unlock()

	filePath := filepath.Join("lang", lang+".json")
	//帮我写个log记录这个filePath是什么
	//fmt.Println("filePath:", filePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var langData map[string]string
	if err := json.Unmarshal(data, &langData); err != nil {
		return err
	}

	translations[lang] = langData
	return nil
}

// T 获取翻译文本
func T(c *gin.Context, key string, args ...interface{}) string {
	mutex.RLock()
	defer mutex.RUnlock()

	var lang string
	if c != nil {
		lang = GetCurrentLang(c)
	} else {
		lang = DefaultLang
	}

	if langData, ok := translations[lang]; ok {
		if text, exists := langData[key]; exists {
			if len(args) > 0 {
				return fmt.Sprintf(text, args...)
			}
			return text
		}
	}
	return key
}

// GetCurrentLang 从上下文获取当前语言设置
func GetCurrentLang(c *gin.Context) string {
	if lang, exists := c.Get(ContextKeyLang); exists {
		return lang.(string)
	}
	return DefaultLang
}

// GetSupportedLanguages 获取支持的语言列表
func GetSupportedLanguages() []string {
	files, err := os.ReadDir("lang")
	if err != nil {
		return []string{"en", "zh"}
	}

	var languages []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			lang := filepath.Base(file.Name())
			languages = append(languages, lang[:len(lang)-5]) // 移除 .json 后缀
		}
	}
	return languages
}

// LanguageMiddleware 语言中间件
func LanguageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetHeader("i18n-next-lng")
		if lang == "" {
			lang = DefaultLang
		}
		// 可以在这里添加语言验证逻辑，确保语言代码是受支持的
		c.Set(ContextKeyLang, lang)
		c.Next()
	}
}
