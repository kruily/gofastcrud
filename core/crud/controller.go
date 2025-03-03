package crud

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kruily/gofastcrud/config"
	"github.com/kruily/gofastcrud/core/crud/module"
	"github.com/kruily/gofastcrud/core/crud/options"
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
	responser := container.MustGetSingletonByName("RESPONSE_HANDLER").(module.ICrudResponse)
	container.BindSingletonWithName(entity.Table(), repo)
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

// buildQueryOptions 构建查询选项
func (c *CrudController[T]) buildQueryOptions(ctx *gin.Context) *options.QueryOptions {
	// 获取基础分页参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", strconv.Itoa(config.CONFIG_MANAGER.GetConfig().Pagenation.DefaultPageSize)))

	// 限制页面大小
	if pageSize > config.CONFIG_MANAGER.GetConfig().Pagenation.MaxPageSize {
		pageSize = config.CONFIG_MANAGER.GetConfig().Pagenation.MaxPageSize
	}

	// 构建查询选项
	opts := options.NewQueryOptions(
		options.WithPage(page),
		options.WithPageSize(pageSize),
		options.WithOrderBy(ctx.DefaultQuery("order_by", "id desc")),
	)
	return opts
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

// standardRoutes 标准路由
func (c *CrudController[T]) standardRoutes(cache bool, cacheTTL int) []types.APIRoute {
	entityName := strings.ToLower(c.entityName[:1]) + c.entityName[1:]
	return []types.APIRoute{
		{
			Path:        "/:" + entityName + "_id",
			Method:      "GET",
			Tags:        []string{c.entityName},
			Summary:     fmt.Sprintf("Get %s by ID", entityName),
			Description: fmt.Sprintf("Get a single %s by its ID", entityName),
			Handler:     c.GetById,
			Response:    c.entity,
			Cache:       types.Cache{Enable: cache, Key: fmt.Sprintf("%s:%s", c.entityName, "getById"), TTL: cacheTTL},
		},
		{
			Path:        "",
			Method:      "GET",
			Tags:        []string{c.entityName},
			Summary:     fmt.Sprintf("List %s", entityName),
			Description: fmt.Sprintf("Get a list of %s with pagination and filters", entityName),
			Handler:     c.List,
			Response:    []T{},
			Parameters:  c.queryParams(),
			Cache:       types.Cache{Enable: cache, Key: fmt.Sprintf("%s:%s", c.entityName, "list"), TTL: cacheTTL},
		},
		{
			Path:        "",
			Method:      "POST",
			Tags:        []string{c.entityName},
			Summary:     fmt.Sprintf("Create %s", entityName),
			Description: fmt.Sprintf("Create a new %s", entityName),
			Handler:     c.Create,
			Request:     c.entity,
			Response:    c.entity,
			Cache:       types.Cache{Enable: cache, Key: fmt.Sprintf("%s:%s", c.entityName, "create"), TTL: cacheTTL},
		},
		{
			Path:        "/:" + entityName + "_id",
			Method:      "POST",
			Tags:        []string{c.entityName},
			Summary:     fmt.Sprintf("Update %s", entityName),
			Description: fmt.Sprintf("Update an existing %s", entityName),
			Handler:     c.Update,
			Request:     c.entity,
			Response:    c.entity,
			Cache:       types.Cache{Enable: cache, Key: fmt.Sprintf("%s:%s", c.entityName, "update"), TTL: cacheTTL},
		},
		{
			Path:        "/:" + entityName + "_id",
			Method:      "DELETE",
			Tags:        []string{c.entityName},
			Summary:     fmt.Sprintf("Delete %s", entityName),
			Description: fmt.Sprintf("Delete an existing %s", entityName),
			Handler:     c.Delete,
			Cache:       types.Cache{Enable: cache, Key: fmt.Sprintf("%s:%s", c.entityName, "delete"), TTL: cacheTTL},
		},
		{
			Path:        "/batch",
			Method:      "POST",
			Tags:        []string{c.entityName},
			Summary:     fmt.Sprintf("Batch Create %s", entityName),
			Description: fmt.Sprintf("Create multiple %s records", entityName),
			Handler:     c.BatchCreate,
			Request:     []T{},
			Response:    []T{},
			Cache:       types.Cache{Enable: cache, Key: fmt.Sprintf("%s:%s", c.entityName, "batchCreate"), TTL: cacheTTL},
		},
		{
			Path:        "/batch",
			Method:      "PUT",
			Tags:        []string{c.entityName},
			Summary:     fmt.Sprintf("Batch Update %s", c.entityName),
			Description: fmt.Sprintf("Update multiple %s records", c.entityName),
			Handler:     c.BatchUpdate,
			Request:     []T{c.entity},
			Response:    []T{c.entity},
			Cache:       types.Cache{Enable: cache, Key: fmt.Sprintf("%s:%s", c.entityName, "batchUpdate"), TTL: cacheTTL},
		},
		{
			Path:        "/batch",
			Method:      "DELETE",
			Tags:        []string{c.entityName},
			Summary:     fmt.Sprintf("Batch Delete %s", entityName),
			Description: fmt.Sprintf("Delete multiple %s records", entityName),
			Handler:     c.BatchDelete,
			Request:     []uint{},
			Response:    nil,
			Cache:       types.Cache{Enable: cache, Key: fmt.Sprintf("%s:%s", c.entityName, "batchDelete"), TTL: cacheTTL},
		},
	}
}

// queryParams 获取查询参数
func (c *CrudController[T]) queryParams() []types.Parameter {
	// 获取实体类型的字段
	entityType := reflect.TypeOf(c.entity)
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	// 收集所有可查询字段
	queryFields := make([]string, 0)
	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)
		// 跳过非导出字段和特殊字段
		if !field.IsExported() || field.Anonymous ||
			field.Type.Kind() == reflect.Struct ||
			field.Type.Kind() == reflect.Slice ||
			field.Type.Kind() == reflect.Map ||
			field.Type.Kind() == reflect.Ptr {
			continue
		}
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}
		// 获取字段名（优先使用json tag）
		fieldName := field.Name
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
		}
		queryFields = append(queryFields, fieldName)
	}

	// 生成查询参数
	queryParams := []types.Parameter{
		{
			Name:        "page",
			In:          "query",
			Description: "Page number",
			Schema:      types.Schema{Type: "integer", Default: "1"},
		},
		{
			Name:        "page_size",
			In:          "query",
			Description: "Number of items per page",
			Schema:      types.Schema{Type: "integer", Default: "10"},
		},
		{
			Name:        "order_by",
			In:          "query",
			Description: "Order by field (e.g., id desc, name asc)",
			Schema:      types.Schema{Type: "string", Default: "id desc"},
		},
		{
			Name:        "search",
			In:          "query",
			Description: "Search keyword",
			Schema:      types.Schema{Type: "string"},
		},
		{
			Name:        "search_fields",
			In:          "query",
			Description: "Fields to search in (comma-separated)",
			Schema:      types.Schema{Type: "string", Default: strings.Join(queryFields, ",")},
		},
		{
			Name:        "preload",
			In:          "query",
			Description: "Relations to preload (comma-separated)",
			Schema:      types.Schema{Type: "string"},
		},
		{
			Name:        "fields",
			In:          "query",
			Description: "Fields to select (comma-separated)",
			Schema:      types.Schema{Type: "string"},
		},
	}

	// 为每个字段添加过滤参数
	for _, field := range queryFields {
		// 大于/小于过滤
		queryParams = append(queryParams,
			types.Parameter{
				Name:        field + "_gt",
				In:          "query",
				Description: fmt.Sprintf("Greater than filter for %s, example: %s_gt=10", field, field),
				Schema:      types.Schema{Type: "string"},
			},
			types.Parameter{
				Name:        field + "_lt",
				In:          "query",
				Description: fmt.Sprintf("Less than filter for %s, example: %s_lt=10", field, field),
				Schema:      types.Schema{Type: "string"},
			},
			types.Parameter{
				Name:        field + "_gte",
				In:          "query",
				Description: fmt.Sprintf("Greater than or equal filter for %s, example: %s_gte=10", field, field),
				Schema:      types.Schema{Type: "string"},
			},
			types.Parameter{
				Name:        field + "_lte",
				In:          "query",
				Description: fmt.Sprintf("Less than or equal filter for %s, example: %s_lte=10", field, field),
				Schema:      types.Schema{Type: "string"},
			},
			types.Parameter{
				Name:        field + "_in",
				In:          "query",
				Description: fmt.Sprintf("IN filter for %s (comma-separated values), example: %s_in=1,2,3", field, field),
				Schema:      types.Schema{Type: "string"},
			},
			types.Parameter{
				Name:        field + "_like",
				In:          "query",
				Description: fmt.Sprintf("LIKE filter for %s, example: %s_like=test", field, field),
				Schema:      types.Schema{Type: "string"},
			},
			types.Parameter{
				Name:        field + "_null",
				In:          "query",
				Description: fmt.Sprintf("NULL filter for %s (true|false), example: %s_null=true", field, field),
				Schema:      types.Schema{Type: "string"},
			},
		)
	}

	return queryParams
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
