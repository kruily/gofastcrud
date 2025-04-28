package models

import (
	"github.com/kruily/gofastcrud/core/crud"
)

// User 用户模型
// @Description 用户信息
type User struct {
	*crud.BaseUUIDEntity
	Username string `json:"username" binding:"required" description:"用户名"`
	Email    string `json:"email" binding:"required" gorm:"unique;" description:"邮箱"`
}

func (*User) TableName() string {
	return "users"
}

func (u *User) Init() {
	if u.BaseUUIDEntity == nil {
		u.BaseUUIDEntity = &crud.BaseUUIDEntity{}
	}
}
