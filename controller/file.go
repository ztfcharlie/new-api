package controller

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"one-api/common"
	"one-api/lang"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 目录是否存在
func isDirExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// 计算上传文件的 MD5 值
func getMd5File(f *multipart.FileHeader) (string, error) {
	file, err := f.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := md5.New()
	// 将文件内容拷贝到 hash 计算器
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	// 重置文件指针，以便后续处理
	if _, err := file.Seek(0, 0); err != nil {
		return "", err
	}
	md5Sum := hex.EncodeToString(hash.Sum(nil))
	// 获取文件扩展名
	ext := filepath.Ext(f.Filename)
	newFilename := md5Sum + ext
	return newFilename, nil
}
func getHost(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	domain := c.Request.URL.Host
	fmt.Println("domain", domain)
	if domain == "" {
		domain = c.Request.Host
	}
	return fmt.Sprintf("%s://%s", scheme, domain)
}
func Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	// 检查文件
	if err != nil {
		if err == http.ErrMissingFile {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": lang.T(c, "error.file.not_exist"),
			})
		} else if strings.Contains(err.Error(), "request body too large") {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": lang.T(c, "error.file.too_large"),
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": lang.T(c, "error.file.upload_failed"),
			})
		}
		return
	}
	// 限制类型
	fileMimeAllowed := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/bmp",
		"image/tif",
		"image/tiff",
		"image/svg+xml",
		"image/webp",
	}
	if condition := file.Header.Get("Content-Type"); !common.StringsContains(fileMimeAllowed, condition) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": lang.T(c, "error.file.not_allowed"),
		})
		return
	}
	// 新建目录
	uploadPath := "./public/uploads/"
	t := time.Now()
	date := t.Format("20060102")
	datePath := filepath.Join(uploadPath, date)
	if !isDirExists(datePath) {
		err := os.Mkdir(datePath, 0777)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": lang.T(c, "error.file.mkdir_failed"),
			})
			return
		}
	}
	// 修改文件名
	md5File, err := getMd5File(file)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": lang.T(c, "error.file.change_name_failed"),
		})
		return
	}
	dst := filepath.Join(datePath, md5File)
	// 保存文件
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": lang.T(c, "error.file.upload_failed"),
		})
		return
	}
	// 转换为URL路径
	urlPath := filepath.ToSlash(strings.TrimPrefix(dst, "public"))
	urlPath = getHost(c) + urlPath
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": lang.T(c, "error.file.upload_success"),
		"data": map[string]string{
			"path": urlPath,
			"file": md5File,
		},
	})
}
