package types

import (
	"github.com/gin-gonic/gin"
)

// HandlerFunc 统一的处理函数类型
type HandlerFunc func(ctx *gin.Context) (interface{}, error)

// Parameter API参数
type Parameter struct {
	Name        string `doc:"name"`
	In          string `doc:"in"` // query, path, header, cookie
	Description string `doc:"description,omitempty"`
	Required    bool   `doc:"required,omitempty"`
	Schema      Schema `doc:"schema"`
}

// Schema 参数架构
type Schema struct {
	Type    string      `doc:"type,omitempty"`
	Format  string      `doc:"format,omitempty"`
	Default interface{} `doc:"default,omitempty"`
}

// Cache 缓存配置
type Cache struct {
	Enable bool   `doc:"enable"` // 是否开启缓存
	Key    string `doc:"key"`    // 缓存键
	TTL    int    `doc:"ttl"`    // 缓存过期时间（秒）
	Force  bool   `doc:"force"`  // 是否强制更新缓存
}

// APIRoute API 路由注解
type APIRoute struct {
	Path        string            `doc:"path"`        // 路径
	Method      string            `doc:"method"`      // HTTP 方法
	Tags        []string          `doc:"tags"`        // 标签分组
	Summary     string            `doc:"summary"`     // 摘要
	Description string            `doc:"description"` // 描述
	Parameters  []Parameter       `doc:"parameters"`  // 参数
	Request     interface{}       `doc:"request"`     // 请求结构体
	Response    interface{}       `doc:"response"`    // 响应结构体
	Handler     HandlerFunc       `doc:"handler"`     // 处理函数
	Middlewares []gin.HandlerFunc `doc:"middlewares"` // 中间件
	Cache       Cache             `doc:"cache"`       // 缓存配置
}

func Post(path string, handler HandlerFunc) APIRoute {
	return APIRoute{
		Path:    path,
		Method:  "POST",
		Handler: handler,
	}
}

func Get(path string, handler HandlerFunc) APIRoute {
	return APIRoute{
		Path:    path,
		Method:  "GET",
		Handler: handler,
	}
}

func Put(path string, handler HandlerFunc) APIRoute {
	return APIRoute{
		Path:    path,
		Method:  "PUT",
		Handler: handler,
	}
}

func Delete(path string, handler HandlerFunc) APIRoute {
	return APIRoute{
		Path:    path,
		Method:  "DELETE",
		Handler: handler,
	}
}

func (r APIRoute) WithSummary(summary string) APIRoute {
	r.Summary = summary
	return r
}

func (r APIRoute) WithTags(tags []string) APIRoute {
	r.Tags = tags
	return r
}

func (r APIRoute) WithDescription(description string) APIRoute {
	r.Description = description
	return r
}

func (r APIRoute) WithParameters(params ...Parameter) APIRoute {
	r.Parameters = params
	return r
}

func (r APIRoute) WithRequest(req interface{}) APIRoute {
	r.Request = req
	return r
}

func (r APIRoute) WithResponse(res interface{}) APIRoute {
	r.Response = res
	return r
}

func (r APIRoute) WithMiddlewares(middlewares ...gin.HandlerFunc) APIRoute {
	r.Middlewares = middlewares
	return r
}

func (r APIRoute) WithCache(enable bool, key string, ttl int, force bool) APIRoute {
	r.Cache = Cache{
		Enable: enable,
		Key:    key,
		TTL:    ttl,
		Force:  force,
	}
	return r
}
