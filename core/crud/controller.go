package crud

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kruily/gofastcrud/core/crud/module"
	"github.com/kruily/gofastcrud/core/crud/types"
	"github.com/kruily/gofastcrud/core/di"
	"github.com/kruily/gofastcrud/errors"
	"gorm.io/gorm"
)

// ICrudController 控制器接口
type ICrudController[T ICrudEntity] interface {
	// 基础 CRUD 操作
	Create(ctx *gin.Context) (interface{}, error)
	GetById(ctx *gin.Context) (interface{}, error)
	Update(ctx *gin.Context) (interface{}, error)
	Delete(ctx *gin.Context) (interface{}, error)
	List(ctx *gin.Context) (interface{}, error)

	// 中间件管理
	UseMiddleware(method string, middlewares ...gin.HandlerFunc)
	GetMiddlewares() map[string][]gin.HandlerFunc

	// 路由注册
	RegisterRoutes()
	GetGroup() *gin.RouterGroup
	SetGroup(group *gin.RouterGroup)
	// GetEntity 获取实体
	GetEntity() ICrudEntity
	// GetEntityName 获取实体名称
	GetEntityName() string
	// EnableCache 启用缓存
	EnableCache(cacheTTL int)

	// 批量操作
	BatchCreate(ctx *gin.Context) (interface{}, error)
	BatchUpdate(ctx *gin.Context) (interface{}, error)
	BatchDelete(ctx *gin.Context) (interface{}, error)
}

// CrudController 控制器实现
type CrudController[T ICrudEntity] struct {
	Repository  IRepository[T]
	Responser   module.ICrudResponse
	Cache       module.ICache
	entity      T
	entityName  string // 添加实体名称字段
	middlewares map[string][]gin.HandlerFunc
	routes      []types.APIRoute
	group       *gin.RouterGroup
}

// NewCrudController 创建控制器
func NewCrudController[T ICrudEntity](db *gorm.DB, entity T) *CrudController[T] {
	entityType := reflect.TypeOf(entity)
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}
	entityName := entityType.Name()
	container := di.SINGLE()
	repo := NewRepository(db, entity)
	responser := container.MustGetSingletonByName(module.ResponseService).(module.ICrudResponse)
	container.BindSingletonWithName(entity.TableName(), repo)
	c := &CrudController[T]{
		Repository:  repo,
		Responser:   responser,
		entity:      entity,
		entityName:  entityName, // 保存实体名称
		middlewares: make(map[string][]gin.HandlerFunc),
	}
	c.routes = append(c.routes, c.standardRoutes(false, 0)...)

	// 自动配置预加载
	c.configurePreloads()

	return c
}

// EnableCache 启用缓存
func (c *CrudController[T]) EnableCache(cacheTTL int) {
	c.routes = []types.APIRoute{}
	c.routes = append(c.routes, c.standardRoutes(true, cacheTTL)...)
}

// configurePreloads 配置预加载
func (c *CrudController[T]) configurePreloads() {
	// 获取实体类型
	entityType := reflect.TypeOf(c.entity)
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	var preloadFields []string
	// 遍历所有字段
	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)

		// 检查是否是关联字段（指针或切片类型）
		if (field.Type.Kind() == reflect.Ptr || field.Type.Kind() == reflect.Slice) &&
			field.Tag.Get("gorm") != "" {
			// 只有带有gorm标签的字段才添加预加载
			preloadFields = append(preloadFields, field.Name)
		}
	}

	// 添加预加载钩子
	if len(preloadFields) > 0 {
		c.Repository.Preload(preloadFields...)
	}
}

// UseMiddleware 添加中间件
func (c *CrudController[T]) UseMiddleware(method string, middlewares ...gin.HandlerFunc) {
	c.middlewares[method] = append(c.middlewares[method], middlewares...)
}

// GetMiddlewares 获取中间件
func (c *CrudController[T]) GetMiddlewares() map[string][]gin.HandlerFunc {
	return c.middlewares
}

// WrapHandler 包装处理函数（公开方法）
func (c *CrudController[T]) WrapHandler(handler types.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 检查请求的APiRoute是否开启缓存
		route := c.GetRoute(ctx)
		if route != nil && route.Cache.Enable {
			// 开启缓存
		}
		result, err := handler(ctx)
		if err != nil {
			var appErr *errors.AppError
			switch e := err.(type) {
			case *errors.AppError:
				appErr = e
			default:
				appErr = errors.Wrap(err, errors.ErrInternal, "内部服务器错误")
			}
			ctx.JSON(appErr.HTTPStatus(), c.Responser.Error(appErr))
			return
		}
		ctx.JSON(200, result)
	}
}

// AddRoute 添加自定义路由
func (c *CrudController[T]) AddRoute(route types.APIRoute) {
	c.routes = append(c.routes, route)
}

// AddRoutes 添加多个自定义路由
func (c *CrudController[T]) AddRoutes(routes []types.APIRoute) {
	c.routes = append(c.routes, routes...)
}

// RegisterRoutes 注册所有路由
func (c *CrudController[T]) RegisterRoutes() {
	// 注册所有路由
	for _, route := range c.routes {
		// 获取中间件
		handlers := c.middlewares["*"]
		handlers = append(handlers, c.middlewares[route.Method]...)
		handlers = append(handlers, route.Middlewares...)
		handlers = append(handlers, c.WrapHandler(route.Handler))

		switch route.Method {
		case "GET":
			c.group.GET(route.Path, handlers...)
		case "POST":
			c.group.POST(route.Path, handlers...)
		case "PUT":
			c.group.PUT(route.Path, handlers...)
		case "DELETE":
			c.group.DELETE(route.Path, handlers...)
		case "PATCH":
			c.group.PATCH(route.Path, handlers...)
		case "OPTIONS":
			c.group.OPTIONS(route.Path, handlers...)
		}
	}
}

// GetRoute 根据请求获取APIRoute
func (c *CrudController[T]) GetRoute(ctx *gin.Context) *types.APIRoute {
	// 获取当前请求的方法和路径
	method := ctx.Request.Method
	path := ctx.Request.URL.Path

	// 获取基础路径和请求路径
	basePath := ctx.FullPath() // 例如: /api/v1/users/:id
	requestPath := path        // 例如: /api/v1/users/123

	for _, route := range c.routes {
		// 方法必须匹配
		if route.Method != method {
			continue
		}

		// 处理不同类型的路由匹配
		switch {
		case route.Path == "": // 空路径匹配根路由，如 /api/v1/users
			if basePath == requestPath {
				return &route
			}
		case route.Path == "/:id": // ID路由匹配，如 /api/v1/users/123
			if len(ctx.Params) > 0 && ctx.Param("id") != "" {
				return &route
			}
		default: // 其他自定义路由
			// 构建完整的路由路径进行比较
			fullRoutePath := basePath[:strings.LastIndex(basePath, "/")] + route.Path
			if fullRoutePath == requestPath {
				return &route
			}
		}
	}

	return nil
}

// GetRoutes 获取所有路由
func (c *CrudController[T]) GetRoutes() []types.APIRoute {
	return c.routes
}

// GetEntityName 获取实体名称
func (c *CrudController[T]) GetEntityName() string {
	return c.entityName
}

// GetEntity 获取实体
func (c *CrudController[T]) GetEntity() ICrudEntity {
	return c.entity
}

// GetGroup 获取路由组
func (c *CrudController[T]) GetGroup() *gin.RouterGroup {
	return c.group
}

// SetGroup 设置路由组
func (c *CrudController[T]) SetGroup(group *gin.RouterGroup) {
	c.group = group
}
