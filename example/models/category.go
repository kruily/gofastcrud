package models

import (
	"github.com/kruily/gofastcrud/core/crud"
)

// Category 分类模型
type Category struct {
	crud.BaseUUIDEntity
	Name  string `json:"name" binding:"required" gorm:"unique"`
	Books []Book `json:"books" gorm:"foreignKey:CategoryID;references:ID"`
}

func (Category) TableName() string {
	return "categories"
}
