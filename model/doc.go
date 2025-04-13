package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Doc struct {
	Id          int            `json:"id"`
	Title       string         `json:"title" gorm:"type:varchar(200);index;not null;comment:标题"`
	Content     string         `json:"content" gorm:"type:text;default:null;comment:文章内容"`
	Keywords    string         `json:"keywords" gorm:"type:varchar(255);default:null;comment:SEO关键词"`
	Description string         `json:"description" gorm:"type:varchar(255);default:null;comment:SEO描述"`
	Summary     string         `json:"summary" gorm:"type:varchar(255);default:null;comment:内容摘要"`
	Views       int            `json:"views" gorm:"type:int(11);default:0;comment:浏览量"`
	Weight      int            `json:"weight" gorm:"type:int(11);default:0;index;comment:权重"`
	Type        int            `json:"type" gorm:"type:tinyint;default:0;index;comment:类型"`
	CreatedAt   time.Time      `json:"created_at" gorm:"bigint;index;comment:创建时间"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"bigint;comment:更新时间"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func GetAllDocs(keywords string, startIdx int, num int) (docs []*Doc, total int64, err error) {
	// Start transaction
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get total count within transaction
	countQuery := tx.Unscoped()
	if keywords != "" {
		countQuery = countQuery.Where("title LIKE ?", "%"+keywords+"%")
	}
	err = countQuery.Model(&Doc{}).Count(&total).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	tx = tx.Unscoped().Select("id", "title", "summary")
	if keywords != "" {
		tx = tx.Where("title LIKE ?", "%"+keywords+"%")
	}
	err = tx.Order("weight desc").Order("created_at desc").Limit(num).Offset(startIdx).Find(&docs).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// Commit transaction
	if err = tx.Commit().Error; err != nil {
		return nil, 0, err
	}

	return docs, total, nil
}
func GetDocById(id int) (*Doc, error) {
	if id == 0 {
		return nil, errors.New("id 为空！")
	}
	doc := Doc{Id: id}
	var err error = nil
	err = DB.First(&doc, "id = ?", id).Error
	return &doc, err
}
