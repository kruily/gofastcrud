package app

import (
	"github.com/kruily/gofastcrud/core/crud/module"
	"github.com/kruily/gofastcrud/core/database"
	"github.com/kruily/gofastcrud/core/di"
)

// Options 应用选项
type AppOption struct {
}

// Option 应用选项
type Option func(*AppOption)

// WithResponse 设置响应处理
func WithResponse(response module.ICrudResponse) Option {
	return func(o *AppOption) {
		// o.Response = response
		di.SINGLE().BindSingletonWithName(module.ResponseService, response)
	}
}

// WithJwt 设置jwt处理
func WithJwt(jwt module.IJwt) Option {
	return func(o *AppOption) {
		// o.Jwt = jwt
		di.SINGLE().BindSingletonWithName(module.JwtService, jwt)
	}
}

// WithCache 设置缓存处理
func WithCache(cache module.ICache) Option {
	return func(o *AppOption) {
		// o.Cache = cache
		di.SINGLE().BindSingletonWithName(module.CacheService, cache)
	}
}

// WithDatabase 设置数据库处理
func WithDatabase(db *database.Database) Option {
	return func(o *AppOption) {
		di.SINGLE().BindSingletonWithName(module.DatabaseService, db)
	}
}
