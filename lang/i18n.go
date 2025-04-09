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

	// 定义可能的搜索路径
	searchPaths := []string{
		filepath.Join("lang", lang+".json"),             // 当前目录下的 lang
		filepath.Join(".", "lang", lang+".json"),        // 显式当前目录
		filepath.Join("..", "lang", lang+".json"),       // 上级目录
		filepath.Join("..", "..", "lang", lang+".json"), // 上上级目录
		filepath.Join("/app/lang", lang+".json"),        // Docker容器中的标准位置
		filepath.Join("/data/lang", lang+".json"),       // 数据目录
	}

	// 获取可执行文件所在目录
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		// 添加可执行文件目录下的搜索路径
		searchPaths = append(searchPaths,
			filepath.Join(execDir, "lang", lang+".json"),
			filepath.Join(filepath.Dir(execDir), "lang", lang+".json"),
		)
	}

	// 获取工作目录
	workDir, err := os.Getwd()
	if err == nil {
		fmt.Printf("Current working directory: %s\n", workDir)
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

	// 如果环境变量中指定了路径，也尝试加载
	if envPath := os.Getenv("LANG_PATH"); envPath != "" {
		path := filepath.Join(envPath, lang+".json")
		fmt.Printf("Trying to load language file from env path: %s\n", path)

		if data, err := os.ReadFile(path); err == nil {
			var langData map[string]string
			if err := json.Unmarshal(data, &langData); err == nil {
				translations[lang] = langData
				fmt.Printf("Successfully loaded language file from env path: %s\n", path)
				return nil
			}
		}
	}

	// 所有尝试都失败了，返回最后一个错误
	return fmt.Errorf("failed to load language file %s.json from any location. Last error: %v", lang, lastErr)
}

// 添加一个初始化函数
func init() {
	// 预加载语言文件
	for _, lang := range []string{"en", "zh"} {
		if err := LoadTranslations(lang); err != nil {
			fmt.Printf("Warning: Failed to load %s translations: %v\n", lang, err)
		}
	}
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

	// 打印当前使用的语言
	fmt.Println("Current language:", lang)

	// 打印是否找到翻译
	//translation := translations[lang][key]
	//fmt.Println("Translation found for key", key, ":", translation != "")

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
