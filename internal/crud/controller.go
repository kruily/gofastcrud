package crud

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kruily/GoFastCrud/internal/crud/types"
	"github.com/kruily/GoFastCrud/pkg/validator"
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
	RegisterRoutes(group *gin.RouterGroup)
}

// CrudController 控制器实现
type CrudController[T ICrudEntity] struct {
	repository  IRepository[T]
	config      *CrudConfig
	entity      T
	entityName  string // 添加实体名称字段
	middlewares map[string][]gin.HandlerFunc
	routes      []types.APIRoute
}

// NewCrudController 创建控制器
func NewCrudController[T ICrudEntity](db *gorm.DB, entity T) *CrudController[T] {
	entityType := reflect.TypeOf(entity)
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}
	entityName := entityType.Name()

	c := &CrudController[T]{
		repository:  NewRepository(db, entity),
		config:      GetConfig(),
		entity:      entity,
		entityName:  entityName, // 保存实体名称
		middlewares: make(map[string][]gin.HandlerFunc),
	}
	c.routes = append(c.routes, c.standardRoutes()...)
	return c
}

// Create 创建实体
func (c *CrudController[T]) Create(ctx *gin.Context) (interface{}, error) {
	var entity T
	if err := ctx.ShouldBindJSON(&entity); err != nil {
		return nil, err
	}

	// 验证实体
	if err := validator.Validate(entity); err != nil {
		return nil, err
	}

	err := c.repository.Create(ctx, &entity)
	if err != nil {
		return nil, err
	}

	return c.config.Responser.Success(entity), nil
}

// GetById 根据ID获取实体
func (c *CrudController[T]) GetById(ctx *gin.Context) (interface{}, error) {
	id := ctx.Param("id")
	if id == "" {
		return nil, errors.New("missing id parameter")
	}

	idInt, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, errors.New("invalid id format")
	}

	entity, err := c.repository.FindById(ctx, uint(idInt))
	if err != nil {
		return nil, err
	}

	if entity == nil {
		return nil, errors.New("record not found")
	}

	return c.config.Responser.Success(entity), nil
}

// List 获取实体列表
func (c *CrudController[T]) List(ctx *gin.Context) (interface{}, error) {
	// 获取查询参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", strconv.Itoa(c.config.DefaultPageSize)))

	// 限制页面大小
	if pageSize > c.config.MaxPageSize {
		pageSize = c.config.MaxPageSize
	}

	// 构建查询选项
	opts := QueryOptions{
		Page:     page,
		PageSize: pageSize,
		OrderBy:  []string{ctx.DefaultQuery("order_by", "id desc")},
	}

	// 执行查询
	items, err := c.repository.Find(ctx, &c.entity, opts)
	if err != nil {
		return nil, err
	}

	// 获取总数
	total, err := c.repository.Count(ctx, &c.entity)
	if err != nil {
		return nil, err
	}

	return c.config.Responser.List(items, total), nil
}

// Update 更新实体
func (c *CrudController[T]) Update(ctx *gin.Context) (interface{}, error) {
	id := ctx.Param("id")
	if id == "" {
		return nil, errors.New("missing id parameter")
	}

	var entity T
	if err := ctx.ShouldBindJSON(&entity); err != nil {
		return nil, err
	}

	// 验证实体
	if err := validator.Validate(entity); err != nil {
		return nil, err
	}

	idInt, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, errors.New("invalid id format")
	}

	entity.SetID(uint(idInt))

	if err := c.repository.Update(ctx, &entity); err != nil {
		return nil, err
	}

	return c.config.Responser.Success(entity), nil
}

// Delete 删除实体
func (c *CrudController[T]) Delete(ctx *gin.Context) (interface{}, error) {
	id := ctx.Param("id")
	if id == "" {
		return nil, errors.New("missing id parameter")
	}

	idInt, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, errors.New("invalid id format")
	}

	opts := []DeleteOptions{{Force: !c.config.SoftDelete}}
	if err := c.repository.DeleteById(ctx, uint(idInt), opts...); err != nil {
		return nil, err
	}

	return c.config.Responser.Success(nil), nil
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
		result, err := handler(ctx)
		if err != nil {
			ctx.JSON(400, c.config.Responser.Error(err))
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
func (c *CrudController[T]) RegisterRoutes(group *gin.RouterGroup) {
	// 注册所有路由
	for _, route := range c.routes {
		// 获取中间件
		handlers := c.middlewares["*"]
		handlers = append(handlers, c.middlewares[route.Method]...)
		handlers = append(handlers, route.Middlewares...)
		handlers = append(handlers, c.WrapHandler(route.Handler))

		switch route.Method {
		case "GET":
			group.GET(route.Path, handlers...)
		case "POST":
			group.POST(route.Path, handlers...)
		case "PUT":
			group.PUT(route.Path, handlers...)
		case "DELETE":
			group.DELETE(route.Path, handlers...)
		}
	}
}

// GetRoutes 获取所有路由
func (c *CrudController[T]) GetRoutes() []types.APIRoute {
	return c.routes
}

// standardRoutes 标准路由
func (c *CrudController[T]) standardRoutes() []types.APIRoute {
	return []types.APIRoute{
		{
			Path:        "/:id",
			Method:      "GET",
			Tags:        []string{c.entityName},
			Summary:     fmt.Sprintf("Get %s by ID", c.entityName),
			Description: fmt.Sprintf("Get a single %s by its ID", c.entityName),
			Handler:     c.GetById,
			Response:    c.entity, // 添加响应类型
		},
		{
			Path:        "",
			Method:      "GET",
			Tags:        []string{c.entityName},
			Summary:     fmt.Sprintf("List %s", c.entityName),
			Description: fmt.Sprintf("Get a list of %s with pagination", c.entityName),
			Handler:     c.List,
			Response:    []T{}, // 添加响应类型
		},
		{
			Path:        "",
			Method:      "POST",
			Tags:        []string{c.entityName},
			Summary:     fmt.Sprintf("Create %s", c.entityName),
			Description: fmt.Sprintf("Create a new %s", c.entityName),
			Handler:     c.Create,
			Request:     c.entity, // 添加请求类型
			Response:    c.entity, // 添加响应类型
		},
		{
			Path:        "/:id",
			Method:      "POST",
			Tags:        []string{c.entityName},
			Summary:     fmt.Sprintf("Update %s", c.entityName),
			Description: fmt.Sprintf("Update an existing %s", c.entityName),
			Handler:     c.Update,
			Request:     c.entity, // 添加请求类型
			Response:    c.entity, // 添加响应类型
		},
		{
			Path:        "/:id",
			Method:      "DELETE",
			Tags:        []string{c.entityName},
			Summary:     fmt.Sprintf("Delete %s", c.entityName),
			Description: fmt.Sprintf("Delete an existing %s", c.entityName),
			Handler:     c.Delete,
		},
	}
}

// GetEntityName 获取实体名称
func (c *CrudController[T]) GetEntityName() string {
	return c.entityName
}
