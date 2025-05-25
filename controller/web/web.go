package web

import (
	"one-api/common"
	"one-api/lang"

	"github.com/gin-gonic/gin"
)

func WebIndex(cc *gin.Context) {
	c := getWebContext(cc)
	// seo
	meta := map[string]string{
		"title":       lang.T(cc, "doc.seo.title"),
		"keywords":    lang.T(cc, "doc.seo.keywords"),
		"description": lang.T(cc, "doc.seo.description"),
	}
	c.render("index.html", gin.H{
		"title":   "",
		"meta":    meta,
		"content": common.OptionMap["HomePageContent"],
	})
}
func WebFaq(cc *gin.Context) {
	c := getWebContext(cc)
	// seo
	meta := map[string]string{
		"title":       lang.T(cc, "doc.seo.title"),
		"keywords":    lang.T(cc, "doc.seo.keywords"),
		"description": lang.T(cc, "doc.seo.description"),
	}
	c.render("faq.html", gin.H{
		"title":   "",
		"meta":    meta,
		"content": common.OptionMap["Faq"],
	})
}
func WebAbout(cc *gin.Context) {
	c := getWebContext(cc)
	// seo
	meta := map[string]string{
		"title":       lang.T(cc, "doc.seo.title"),
		"keywords":    lang.T(cc, "doc.seo.keywords"),
		"description": lang.T(cc, "doc.seo.description"),
	}
	c.render("about.html", gin.H{
		"title":   "",
		"meta":    meta,
		"content": common.OptionMap["About"],
	})
}
