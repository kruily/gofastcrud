package models

import "github.com/kruily/GoFastCrud/internal/crud"

// Book 书籍模型
type Book struct {
	*crud.BaseEntity `json:",inline"`
	Title            string `json:"title" example:"The Great Gatsby" description:"书名"`
	Author           string `json:"author" example:"F. Scott Fitzgerald" description:"作者"`
}

func (Book) Table() string {
	return "books"
}
