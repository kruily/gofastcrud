package models

import (
	"github.com/kruily/GoFastCrud/internal/crud"
)

// User 用户模型
// @Description 用户信息
type User struct {
	*crud.BaseEntity `json:",inline"`
	Username         string `json:"username" example:"john_doe" description:"用户名"`
	Email            string `json:"email" example:"john@example.com" description:"电子邮件"`
}

func (User) Table() string {
	return "users"
}
