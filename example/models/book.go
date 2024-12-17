package models

import (
	"github.com/kruily/gofastcrud/core/crud"
)

// Book 书籍模型
type Book struct {
	*crud.BaseUUIDEntity
	Title      string    `json:"title" binding:"required"`
	CategoryID string    `json:"category_id" gorm:"type:text;index:idx_category_id(255)"`
	Category   *Category `json:"category" gorm:"foreignKey:CategoryID;references:ID"`
}

func (Book) Table() string {
	return "books"
}
