package types

import "github.com/gin-gonic/gin"

// HandlerFunc 统一的处理函数类型
type HandlerFunc func(ctx *gin.Context) (interface{}, error)

// APIRoute API 路由注解
type APIRoute struct {
	Path        string            // 路径
	Method      string            // HTTP 方法
	Tags        []string          // 标签分组
	Summary     string            `doc:"summary"`     // 摘要
	Description string            `doc:"description"` // 描述
	Request     interface{}       // 请求结构体
	Response    interface{}       // 响应结构体
	Handler     HandlerFunc       // 处理函数
	Middlewares []gin.HandlerFunc // 中间件
}
