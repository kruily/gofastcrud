package types

import "github.com/gin-gonic/gin"

// APIVersion 版本类型
type APIVersion string

// RouteRegister 路由注册函数类型
type RouteRegister func(r *gin.RouterGroup)
