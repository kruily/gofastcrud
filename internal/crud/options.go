package crud

import (
	"time"

	"gorm.io/gorm"
)

// QueryOptions 查询选项
type QueryOptions struct {
	// 分页
	Page     int
	PageSize int
	// 排序
	OrderBy []string
	// 查询条件
	Where map[string]interface{}
	// 预加载关系
	Preload []string
	// 选择特定字段
	Select []string
	// 搜索关键词
	Search string
	// 搜索字段
	SearchFields []string
	// 过滤条件
	Filter map[string]interface{}
}

// QueryHook 查询钩子
type QueryHook interface {
	BeforeQuery(*gorm.DB)
	AfterQuery(*gorm.DB)
}

// DeleteOptions 删除选项
type DeleteOptions struct {
	// 是否物理删除
	Force bool
	// 删除时间
	DeletedAt time.Time
	// 删除人
	DeletedBy string
}

// BatchOptions 批量操作选项
type BatchOptions struct {
	BatchSize int  // 每批次处理数量
	Async     bool // 是否异步处理
}
