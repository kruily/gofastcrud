package models

import "github.com/kruily/GoFastCrud/internal/crud"

// Book 书籍模型
type Book struct {
	*crud.BaseEntity `json:",inline"`
	Title            string    `json:"title" binding:"required"`
	CategoryID       uint      `json:"category_id"`
	Category         *Category `json:"category" gorm:"foreignKey:CategoryID;references:ID"`
}

func (Book) Table() string {
	return "books"
}
