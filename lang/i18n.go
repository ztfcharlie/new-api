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
	ContextKeyLang = "current_language"
	DefaultLang    = "en"
)

var (
	translations = make(map[string]map[string]string)
	mutex        sync.RWMutex
)

// LoadTranslations 加载指定语言的翻译文件
func LoadTranslations(lang string) error {
	mutex.Lock()
	defer mutex.Unlock()

	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return err
	}

	filePath := filepath.Join("lang", lang+".json")
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		fmt.Printf("Error getting absolute path: %v\n", err)
		return err
	}

	// 打印详细的路径信息
	fmt.Printf("Current Directory: %s\n", currentDir)
	fmt.Printf("Relative Path: %s\n", filePath)
	fmt.Printf("Absolute Path: %s\n", absPath)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("Translation file does not exist: %s\n", filePath)
		return err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading translation file: %v\n", err)
		return err
	}

	// 打印文件内容长度和部分内容
	fmt.Printf("File content length: %d bytes\n", len(data))
	if len(data) > 100 {
		fmt.Printf("First 100 bytes: %s\n", string(data[:100]))
	}

	var langData map[string]string
	if err := json.Unmarshal(data, &langData); err != nil {
		fmt.Printf("Error unmarshaling JSON: %v\n", err)
		return err
	}

	// 打印加载的翻译数量
	fmt.Printf("Loaded %d translations for language: %s\n", len(langData), lang)

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

	fmt.Printf("Looking up translation - Language: %s, Key: %s\n", lang, key)

	// 检查语言数据是否存在
	langData, ok := translations[lang]
	if !ok {
		fmt.Printf("No translations found for language: %s\n", lang)
		return key
	}

	// 检查翻译是否存在
	text, exists := langData[key]
	if !exists {
		fmt.Printf("No translation found for key: %s in language: %s\n", key, lang)
		return key
	}

	if len(args) > 0 {
		result := fmt.Sprintf(text, args...)
		fmt.Printf("Formatted translation: %s\n", result)
		return result
	}

	fmt.Printf("Found translation: %s\n", text)
	return text
}

// GetCurrentLang 从上下文获取当前语言设置
func GetCurrentLang(c *gin.Context) string {
	if lang, exists := c.Get(ContextKeyLang); exists {
		fmt.Printf("Language from context: %s\n", lang)
		return lang.(string)
	}
	fmt.Printf("Using default language: %s\n", DefaultLang)
	return DefaultLang
}

// GetSupportedLanguages 获取支持的语言列表
func GetSupportedLanguages() []string {
	files, err := os.ReadDir("lang")
	if err != nil {
		fmt.Printf("Error reading lang directory: %v\n", err)
		return []string{"en", "zh"}
	}

	var languages []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			lang := filepath.Base(file.Name())
			languages = append(languages, lang[:len(lang)-5])
		}
	}
	fmt.Printf("Supported languages: %v\n", languages)
	return languages
}

// LanguageMiddleware 语言中间件
func LanguageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetHeader("i18n-next-lng")
		if lang == "" {
			lang = DefaultLang
		}
		fmt.Printf("Language middleware - Selected language: %s\n", lang)
		c.Set(ContextKeyLang, lang)
		c.Next()
	}
}
