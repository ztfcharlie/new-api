package router

import (
	"one-api/controller/web"
	"one-api/middleware"
	"one-api/views"

	"github.com/gin-gonic/gin"
)

func SetWebRenderRouter(router *gin.Engine) {
	// router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.SetHTMLTemplate(views.Templates)
	webRouter := router.Group("")
	{
		webRouter.Use(middleware.GlobalWebRateLimit())
		// 静态资源
		webRouter.Static("/static", "./public/static")
		// webRouter.Static("/webHtml", "./public/webHtml")
		webRouter.Static("/uploads", "./public/uploads")
		// // 模板目录
		// webRouter.GET("/docs", web.GetAllDocs)
		// webRouter.GET("/doc/:slug", web.GetDoc)
	}
	ssrRouter := router.Group("/ssr")
	{
		ssrRouter.GET("/", web.WebIndex)
		ssrRouter.GET("/faq", web.WebFaq)
		ssrRouter.GET("/about", web.WebAbout)

		// 模板目录
		ssrRouter.GET("/docs", web.GetAllDocs)
		ssrRouter.GET("/doc/:slug", web.GetDoc)
	}

}
