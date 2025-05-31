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
		"title":       lang.T(cc, "doc.seo.title.index"),
		"keywords":    lang.T(cc, "doc.seo.keywords.index"),
		"description": lang.T(cc, "doc.seo.description.index"),
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
		"title":       lang.T(cc, "doc.seo.title.WebFaq"),
		"keywords":    lang.T(cc, "doc.seo.keywords.WebFaq"),
		"description": lang.T(cc, "doc.seo.description.WebFaq"),
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
		"title":       lang.T(cc, "doc.seo.title.WebAbout"),
		"keywords":    lang.T(cc, "doc.seo.keywords_WebAbout"),
		"description": lang.T(cc, "doc.seo.description.WebAbout"),
	}
	c.render("about.html", gin.H{
		"title":   "",
		"meta":    meta,
		"content": common.OptionMap["About"],
	})
}
