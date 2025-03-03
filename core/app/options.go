package app

import (
	"github.com/kruily/gofastcrud/core/crud/module"
)

// Options 应用选项
type AppOption struct {
	Response module.ICrudResponse
}

// Option 应用选项
type Option func(*AppOption)

// WithResponse 设置响应处理
func WithResponse(response module.ICrudResponse) Option {
	return func(o *AppOption) {
		o.Response = response
	}
}
