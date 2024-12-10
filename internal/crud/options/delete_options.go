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

// NewDeleteOptions 创建删除选项
func NewDeleteOptions(opts ...*DeleteOptions) *DeleteOptions {
	opt := &DeleteOptions{}
	for _, o := range opts {
		*opt = *o
	}
	return opt
}

// WithForce 设置是否物理删除
func WithForce(force bool) *DeleteOptions {
	return &DeleteOptions{Force: force}
}

// WithDeletedAt 设置删除时间
func WithDeletedAt(deletedAt time.Time) *DeleteOptions {
	return &DeleteOptions{DeletedAt: deletedAt}
}

// WithDeletedBy 设置删除人
func WithDeletedBy(deletedBy string) *DeleteOptions {
	return &DeleteOptions{DeletedBy: deletedBy}
}
