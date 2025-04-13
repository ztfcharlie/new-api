package web

import (
	"html/template"
	"net/http"
	"one-api/model"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetDoc(cc *gin.Context) {
	c := getWebContext(cc)

	id, _ := strconv.Atoi(c.Param("id"))
	doc, err := model.GetDocById(id)
	if err != nil {
		c.String(http.StatusNotFound, "404")
		return
	}
	// seo
	meta := map[string]string{
		"title":       doc.Title,
		"keywords":    doc.Keywords,
		"description": doc.Description,
	}
	c.render("doc-detail.html", gin.H{
		"meta":       meta,
		"doc":        doc,
		"docContent": template.HTML(doc.Content),
	})

}

func GetAllDocs(cc *gin.Context) {
	c := getWebContext(cc)
	keywords := c.Query("keywords")
	p, startIdx, pageSize := c.getPageParams()
	docs, total, err := model.GetAllDocs(keywords, startIdx, pageSize)
	if err != nil {
		c.abortError(http.StatusOK, err.Error())
		return
	}
	// seo
	meta := map[string]string{
		"title":       "文章列表",
		"keywords":    "文章,doc",
		"description": "文章描述",
	}
	c.render("doc.html", gin.H{
		"keywords":   keywords,
		"docs":       docs,
		"meta":       meta,
		"pagination": c.getPagination(p, pageSize, total),
	})
}
