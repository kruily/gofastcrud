package models

import "github.com/kruily/GoFastCrud/internal/crud"

// Category 分类模型
type Category struct {
	*crud.BaseEntity `json:",inline"`
	Name             string `json:"name" binding:"required" gorm:"unique"`
	Books            []Book `json:"books" gorm:"foreignKey:CategoryID;references:ID"`
}

func (Category) Table() string {
	return "categories"
}
