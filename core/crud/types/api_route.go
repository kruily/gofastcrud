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
