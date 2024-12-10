package models

import "github.com/kruily/gofastcrud/core/crud"

// User 用户模型
// @Description 用户信息
type User struct {
	*crud.BaseEntity
	Username string `json:"username" binding:"required"`
	Email    string `json:"email"`
}

func (User) Table() string {
	return "users"
}
