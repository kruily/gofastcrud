package options

import "time"

// DeleteOptions 删除选项
type DeleteOptions struct {
	// 是否物理删除
	Force bool
	// 删除时间
	DeletedAt time.Time
	// 删除人
	DeletedBy string
}
