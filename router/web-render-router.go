package router

import (
	"one-api/controller/web"
	"one-api/middleware"
	"one-api/views"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func SetWebRenderRouter(router *gin.Engine) {
	// router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.SetHTMLTemplate(views.Templates)
	webRouter := router.Group("")
	{
		webRouter.Use(middleware.GlobalWebRateLimit())

		exePath, _ := os.Executable()
		// 获取可执行文件所在目录
		exeDir := filepath.Dir(exePath)
		// 静态资源
		webRouter.Static("/static", filepath.Join(exeDir, "./public/static"))
		webRouter.Static("/webHtml", filepath.Join(exeDir, "./public/webHtml"))
		// 模板目录
		webRouter.GET("/docs", web.GetAllDocs)
		webRouter.GET("/doc/:id", web.GetDoc)

	}

}
