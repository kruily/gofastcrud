package models

import "github.com/kruily/GoFastCrud/internal/crud"

// User 用户模型
// @Description 用户信息
type User struct {
	*crud.BaseEntity
	ID       uint   `json:"id" gorm:"primarykey"`
	Username string `json:"username" binding:"required" validate:"required,min=3,max=32"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

func (User) Table() string {
	return "users"
}
