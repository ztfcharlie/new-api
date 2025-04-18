package web

import (
	"html/template"
	"net/http"
	"one-api/lang"
	"one-api/model"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetDoc(cc *gin.Context) {
	c := getWebContext(cc)
	id, _ := strconv.Atoi(c.Param("id"))
	tmpDoc := model.Doc{Id: id, Status: 1}
	doc, err := tmpDoc.GetDoc()
	if err != nil {
		c.abortError(http.StatusNotFound, lang.T(nil, "error.status.404"))
		return
	}
	doc.Increment("views")
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
	query := model.DocQuery{
		Title:  keywords,
		Status: 1,
	}
	docs, total, err := model.GetAllDocs(query, startIdx, pageSize)
	if err != nil {
		c.abortError(http.StatusOK, err.Error())
		return
	}

	// seo
	meta := map[string]string{
		"title":       lang.T(cc, "doc.seo.title"),
		"keywords":    lang.T(cc, "doc.seo.keywords"),
		"description": lang.T(cc, "doc.seo.description"),
	}
	c.render("doc.html", gin.H{
		"title":      keywords,
		"docs":       docs,
		"meta":       meta,
		"pagination": c.getPagination(p, pageSize, total),
	})
}
