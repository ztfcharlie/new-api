package views

import (
	"embed"
	"html/template"
	"one-api/lang"
	"time"
)

//go:embed */*.html
var templatesFs embed.FS

// 模板函数
var funcMap = template.FuncMap{
	// 输出html
	"safe": func(str string) template.HTML {
		return template.HTML(str)
	},
	// 格式化日期
	"formatDate": func(timeValue time.Time) string {
		return timeValue.Format("02/01 2006")
	},
	// 年份
	"year": func() int {
		return time.Now().Year()
	},
	// 国际化
	"T": func(key string) string {
		return lang.T(nil, key)
	},
}

// 模板文件
var Templates = template.Must(template.New("").Funcs(funcMap).ParseFS(templatesFs, "*/*.html"))
