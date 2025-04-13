package web

import (
	"math"
	"net/url"
	"one-api/common"
	"strconv"
)

// 页面连接
type pageLink struct {
	Link     string // 链接
	Page     int    // 页码
	Disabled bool   // 是否禁用
}

// 分页
type pagination struct {
	StartCount int        // 开始数量
	EndCount   int        // 结束数量
	Total      int        // 总数
	Page       int        // 页码
	PrevPage   pageLink   // 上一页
	NextPage   pageLink   // 下一页
	Pages      []pageLink // 页数集合
}

// 构建分页链接
func buildPageLink(baseURL string, pageNum int) pageLink {
	urlObj, _ := url.Parse(baseURL)
	newQuery := urlObj.Query()
	newQuery.Set("p", strconv.Itoa(pageNum))
	urlObj.RawQuery = newQuery.Encode()
	return pageLink{
		Link:     urlObj.String(),
		Page:     pageNum,
		Disabled: false,
	}
}

// 获取页面参数
func (c *webContext) getPageParams() (int, int, int) {
	p, _ := strconv.Atoi(c.Query("p"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if p < 1 {
		p = 1
	}
	if pageSize <= 0 {
		pageSize = common.ItemsPerPage
	}
	startIdx := (p - 1) * pageSize
	return p, startIdx, pageSize
}

// 获取分页
func (c *webContext) getPagination(page int, pageSize int, total64 int64) pagination {
	// 基本链接
	query := c.Request.URL.Query()
	query.Del("p")
	baseURL := c.Request.URL.Path
	if len(query) > 0 {
		baseURL = baseURL + "?" + query.Encode()
	}
	// 计算当前页数和最大页数
	total := int(total64)
	startCount := (page - 1) * pageSize
	if startCount <= 0 {
		startCount = 1
	}
	endCount := page * pageSize
	if endCount > total {
		endCount = total
	}
	maxPage := int(math.Ceil(float64(total) / float64(pageSize)))

	// 构建上一页和下一页
	var prevPage, nextPage pageLink
	if page > 1 {
		prevPage = buildPageLink(baseURL, page-1)
	} else {
		prevPage = buildPageLink(baseURL, 1)
		prevPage.Disabled = true
	}
	if page < maxPage {
		nextPage = buildPageLink(baseURL, page+1)
	} else {
		nextPage = buildPageLink(baseURL, maxPage)
		nextPage.Disabled = true
	}

	// 构建链接集合
	var pages []pageLink
	for i := 1; i <= maxPage; i++ {
		pages = append(pages, buildPageLink(baseURL, i))
	}
	return pagination{
		StartCount: startCount,
		EndCount:   endCount,
		Total:      total,
		Page:       page,
		PrevPage:   prevPage,
		NextPage:   nextPage,
		Pages:      pages,
	}
}
