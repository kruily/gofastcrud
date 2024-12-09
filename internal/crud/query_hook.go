package crud

import (
	"gorm.io/gorm"
)

// QueryHook 查询钩子
type QueryHook interface {
	BeforeQuery(*gorm.DB)
	AfterQuery(*gorm.DB)
}
