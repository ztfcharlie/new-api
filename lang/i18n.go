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

func init() {
	fmt.Println("=== Language System Debug Info ===")

	// 打印当前工作目录
	if workDir, err := os.Getwd(); err == nil {
		fmt.Printf("Working Directory: %s\n", workDir)
	}

	// 打印可执行文件位置
	if execPath, err := os.Executable(); err == nil {
		fmt.Printf("Executable Path: %s\n", execPath)
	}

	// 列出 /app/lang 目录内容
	if files, err := os.ReadDir("/app/lang"); err == nil {
		fmt.Println("Contents of /app/lang:")
		for _, file := range files {
			fmt.Printf("  - %s\n", file.Name())
		}
	}

	fmt.Println("===========================")
}

// LoadTranslations 加载指定语言的翻译文件
// LoadTranslations 加载指定语言的翻译文件
func LoadTranslations(lang string) error {
	mutex.Lock()
	defer mutex.Unlock()

	// 定义可能的搜索路径
	searchPaths := []string{
		filepath.Join("/app/lang", lang+".json"),  // Docker容器中的标准位置
		filepath.Join("/data/lang", lang+".json"), // 当前工作目录
		filepath.Join("lang", lang+".json"),       // 相对路径
		filepath.Join("..", "lang", lang+".json"), // 上级目录
	}

	// 获取可执行文件所在目录
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		// 添加可执行文件相关的路径
		searchPaths = append(searchPaths,
			filepath.Join(execDir, "lang", lang+".json"),
			filepath.Join(filepath.Dir(execDir), "lang", lang+".json"),
		)
	}

	// 获取当前工作目录
	if workDir, err := os.Getwd(); err == nil {
		fmt.Printf("Current working directory: %s\n", workDir)
		// 添加工作目录相关的路径
		searchPaths = append(searchPaths,
			filepath.Join(workDir, "lang", lang+".json"),
		)
	}

	// 尝试所有可能的路径
	var lastErr error
	for _, path := range searchPaths {
		absPath, _ := filepath.Abs(path)
		fmt.Printf("Trying to load language file from: %s\n", absPath)

		data, err := os.ReadFile(path)
		if err == nil {
			var langData map[string]string
			if err := json.Unmarshal(data, &langData); err != nil {
				fmt.Printf("Error unmarshaling JSON from %s: %v\n", absPath, err)
				lastErr = err
				continue
			}

			translations[lang] = langData
			fmt.Printf("Successfully loaded language file from: %s\n", absPath)
			return nil
		}
		lastErr = err
	}

	return fmt.Errorf("failed to load language file %s.json from any location. Last error: %v", lang, lastErr)
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
