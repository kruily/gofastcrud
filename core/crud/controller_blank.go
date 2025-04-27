package crud

import (
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/kruily/gofastcrud/core/crud/module"
	"github.com/kruily/gofastcrud/core/crud/types"
	"github.com/kruily/gofastcrud/core/di"
	"gorm.io/gorm"
)

type BlankController[T ICrudEntity] struct {
	Repository  IRepository[T]
	Responser   module.ICrudResponse
	Cache       module.ICache
	entity      T
	entityName  string // 添加实体名称字段
	middlewares map[string][]gin.HandlerFunc
	routes      []*types.APIRoute
	group       *gin.RouterGroup
}

// NewCrudController 创建控制器
func NewBlankController[T ICrudEntity](db *gorm.DB, entity T) *BlankController[T] {
	entity.Init()
	entityType := reflect.TypeOf(entity)
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}
	entityName := entityType.Name()
	container := di.SINGLE()
	repo := NewRepository(db, entity)
	responser := container.MustGetSingletonByName(module.ResponseService).(module.ICrudResponse)
	container.BindSingletonWithName(entity.TableName(), repo)
	c := &BlankController[T]{
		Repository:  repo,
		Responser:   responser,
		entity:      entity,
		entityName:  entityName, // 保存实体名称
		middlewares: make(map[string][]gin.HandlerFunc),
		routes:      make([]*types.APIRoute, 0),
	}
	// c.routes = append(c.routes, c.standardRoutes(false, 0)...)

	// 自动配置预加载
	c.configurePreloads()

	return c
}

// configurePreloads 配置预加载
func (c *BlankController[T]) configurePreloads() {
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
func (c *BlankController[T]) UseMiddleware(method string, middlewares ...gin.HandlerFunc) {
	c.middlewares[method] = append(c.middlewares[method], middlewares...)
}

// GetMiddlewares 获取中间件
func (c *BlankController[T]) GetMiddlewares() map[string][]gin.HandlerFunc {
	return c.middlewares
}

// AddRoute 添加自定义路由
func (c *BlankController[T]) AddRoute(route *types.APIRoute) {
	c.routes = append(c.routes, route)
}

// AddRoutes 添加多个自定义路由
func (c *BlankController[T]) AddRoutes(routes []*types.APIRoute) {
	c.routes = append(c.routes, routes...)
}

// ClearRoutes 清除所有自定义路由(注册到gin后使用)
func (c *BlankController[T]) ClearRoutes() {
	c.routes = []*types.APIRoute{}
}

// RegisterRoutes 注册所有路由
func (c *BlankController[T]) RegisterRoutes() {
	controllerRegisterRoute(c)
}

// GetRoute 根据请求获取APIRoute
func (c *BlankController[T]) GetRoute(ctx *gin.Context) *types.APIRoute {
	return getRoute(c, ctx)
}

// GetRoutes 获取所有路由
func (c *BlankController[T]) GetRoutes() []*types.APIRoute {
	return c.routes
}

// GetEntityName 获取实体名称
func (c *BlankController[T]) GetEntityName() string {
	return c.entityName
}

// GetEntity 获取实体
func (c *BlankController[T]) GetEntity() ICrudEntity {
	return c.entity
}

// GetGroup 获取路由组
func (c *BlankController[T]) GetGroup() *gin.RouterGroup {
	return c.group
}

// SetGroup 设置路由组
func (c *BlankController[T]) SetGroup(group *gin.RouterGroup) {
	c.group = group
}

// GetResponser 获取响应器
func (c *BlankController[T]) GetResponser() module.ICrudResponse {
	return c.Responser
}
