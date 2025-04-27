package app

import (
	"github.com/kruily/gofastcrud/core/crud/module"
)

// Options 应用选项
type AppOption struct {
	Response module.ICrudResponse
	Jwt      module.IJwt
	cache    module.ICache
}

// Option 应用选项
type Option func(*AppOption)

// WithResponse 设置响应处理
func WithResponse(response module.ICrudResponse) Option {
	return func(o *AppOption) {
		o.Response = response
	}
}

// WithJwt 设置jwt处理
func WithJwt(jwt module.IJwt) Option {
	return func(o *AppOption) {
		o.Jwt = jwt
	}
}

// WithCache 设置缓存处理
func WithCache(cache module.ICache) Option {
	return func(o *AppOption) {
		o.cache = cache
	}
}
