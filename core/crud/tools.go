package crud

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kruily/gofastcrud/core/crud/module"
	"github.com/kruily/gofastcrud/core/crud/types"
	"github.com/kruily/gofastcrud/errors"
)

// 泛型函数：创建并初始化 T 的实例
func NewModel[T ICrudEntity]() T {
	var t T
	// 获取 T 的类型信息
	typ := reflect.TypeOf(t)
	// 必须是指针类型（如 *People）
	if typ.Kind() != reflect.Ptr {
		panic("T must be a pointer type")
	}
	// 创建新实例（例如 &People{}）
	instance := reflect.New(typ.Elem()).Interface().(T)
	instance.Init() // 调用初始化方法
	return instance
}

func getRoute(c ICrudController[ICrudEntity], ctx *gin.Context) *types.APIRoute {
	// 获取当前请求的方法和路径
	method := ctx.Request.Method
	path := ctx.Request.URL.Path

	// 获取基础路径和请求路径
	basePath := ctx.FullPath() // 例如: /api/v1/users/:id
	requestPath := path        // 例如: /api/v1/users/123

	for _, route := range c.GetRoutes() {
		// 方法必须匹配
		if route.Method != method {
			continue
		}

		// 处理不同类型的路由匹配
		switch {
		case route.Path == "": // 空路径匹配根路由，如 /api/v1/users
			if basePath == requestPath {
				return route
			}
		case route.Path == "/:id": // ID路由匹配，如 /api/v1/users/123
			if len(ctx.Params) > 0 && ctx.Param("id") != "" {
				return route
			}
		default: // 其他自定义路由
			// 构建完整的路由路径进行比较
			fullRoutePath := basePath[:strings.LastIndex(basePath, "/")] + route.Path
			if fullRoutePath == requestPath {
				return route
			}
		}
	}
	return nil
}

// controllerRegisterRoute 注册所有路由 所有controller基类
func controllerRegisterRoute(c ICrudController[ICrudEntity]) {
	// 注册所有路由
	for _, route := range c.GetRoutes() {
		// 获取中间件
		handlers := c.GetMiddlewares()["*"]
		handlers = append(handlers, c.GetMiddlewares()[route.Method]...)
		handlers = append(handlers, route.Middlewares...)
		handlers = append(handlers, WrapHandler(route.Handler, c.GetResponser()))

		switch route.Method {
		case "GET":
			c.GetGroup().GET(route.Path, handlers...)
		case "POST":
			c.GetGroup().POST(route.Path, handlers...)
		case "PUT":
			c.GetGroup().PUT(route.Path, handlers...)
		case "DELETE":
			c.GetGroup().DELETE(route.Path, handlers...)
		case "PATCH":
			c.GetGroup().PATCH(route.Path, handlers...)
		case "OPTIONS":
			c.GetGroup().OPTIONS(route.Path, handlers...)
		}
	}
}

// WrapHandler 包装处理函数（公开方法）
func WrapHandler(handler types.HandlerFunc, response module.ICrudResponse) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// // 检查请求的APiRoute是否开启缓存
		// route := c.GetRoute(ctx)
		// if route != nil && route.Cache.Enable {
		// 	// 开启缓存
		// }
		// 添加日志记录中间件，记录请求信息
		result, err := handler(ctx)
		if err != nil {
			var appErr *errors.AppError
			switch e := err.(type) {
			case *errors.AppError:
				appErr = e
			default:
				appErr = errors.Wrap(err, errors.ErrInternal, "内部服务器错误")
			}
			ctx.JSON(appErr.HTTPStatus(), response.Error(appErr))
			return
		}
		// 日志记录请求返回结果
		ctx.JSON(200, result)
	}
}
