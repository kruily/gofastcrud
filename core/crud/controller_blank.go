package crud

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kruily/gofastcrud/config"
	"github.com/kruily/gofastcrud/core/crud/module"
	"github.com/kruily/gofastcrud/core/crud/options"
	"github.com/kruily/gofastcrud/core/crud/types"
	"github.com/kruily/gofastcrud/core/di"
	"gorm.io/gorm"
)

// ICrudController 控制器接口
type ICrudController[T ICrudEntity] interface {
	// 中间件管理
	GetMiddlewares() map[string][]gin.HandlerFunc

	// 路由注册
	RegisterRoutes()
	GetGroup() *gin.RouterGroup      // 获取路由组
	SetGroup(group *gin.RouterGroup) // 设置路由组
	GetEntity() ICrudEntity          // GetEntity 获取实体
	GetEntityName() string           // GetEntityName 获取实体名称
	// 路由管理
	AddRoute(route *types.APIRoute)     // 添加自定义路由
	AddRoutes(routes []*types.APIRoute) // 添加多个自定义路由
	GetRoutes() []*types.APIRoute       // 获取所有路由
	ClearRoutes()                       // 清除所有路由
	GetResponser() module.ICrudResponse // 获取响应器
}

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

// queryParams 获取查询参数
func (c *BlankController[T]) queryParams() []types.Parameter {
	// 获取所有可查询字段
	queryFields := queryFields(c.entity)
	searchFields := make([]string, 0)
	for _, field := range queryFields {
		if field.FilterTag != "" {
			searchFields = append(searchFields, field.Field)
		}
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
			Schema:      types.Schema{Type: "string", Default: strings.Join(searchFields, ",")},
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
	queryParams = append(queryParams, ModeParams(c)...)
	return queryParams
}

// BuildQueryOptions 构建查询选项
func (c *BlankController[T]) BuildQueryOptions(ctx *gin.Context) *options.QueryOptions {
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

	// 处理搜索
	if search := ctx.Query("search"); search != "" {
		opts.Search = search
		opts.SearchFields = strings.Split(ctx.DefaultQuery("search_fields", "id"), ",")
	}

	// 处理过滤条件
	c.buildFilterOptions(ctx, opts)

	// 处理预加载关系
	if preload := ctx.Query("preload"); preload != "" {
		opts.Preload = strings.Split(preload, ",")
	}

	// 处理字段选择
	if fields := ctx.Query("fields"); fields != "" {
		opts.Select = strings.Split(fields, ",")
	}

	return opts
}

// buildFilterOptions 构建过滤选项
func (c *BlankController[T]) buildFilterOptions(ctx *gin.Context, opts *options.QueryOptions) {
	querys := ctx.Request.URL.Query()
	for key, values := range querys {
		// 跳过特殊参数
		if isSpecialParam(key) {
			continue
		}
		// 处理范围查询
		if strings.HasSuffix(key, "_gt") || strings.HasSuffix(key, "_lt") ||
			strings.HasSuffix(key, "_gte") || strings.HasSuffix(key, "_lte") {
			field := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(
				strings.TrimSuffix(key, "_gt"), "_lt"), "_gte"), "_lte")

			if strings.HasSuffix(key, "_gt") {
				opts.Where[field+" > ?"] = values[0]
			} else if strings.HasSuffix(key, "_lt") {
				opts.Where[field+" < ?"] = values[0]
			} else if strings.HasSuffix(key, "_gte") {
				opts.Where[field+" >= ?"] = values[0]
			} else if strings.HasSuffix(key, "_lte") {
				opts.Where[field+" <= ?"] = values[0]
			}
			continue
		}

		// 处理IN查询
		if strings.HasSuffix(key, "_in") {
			field := strings.TrimSuffix(key, "_in")
			opts.Where[field+" IN ?"] = strings.Split(values[0], ",")
			continue
		}

		// 处理NULL查询
		if strings.HasSuffix(key, "_null") {
			field := strings.TrimSuffix(key, "_null")
			if values[0] == "true" {
				opts.Where[field+" IS NULL"] = nil
			} else {
				opts.Where[field+" IS NOT NULL"] = nil
			}
			continue
		}

		// 处理模糊查询
		if strings.HasSuffix(key, "_like") {
			field := strings.TrimSuffix(key, "_like")
			opts.Where[field+" LIKE ?"] = "%" + values[0] + "%"
			continue
		}

		// 处理普通相等查询
		if !isSpecialParam(key) {
			opts.Filter[key] = values[0]
		}
	}
}

// isSpecialParam 检查是否为特殊参数
func isSpecialParam(key string) bool {
	specialParams := []string{
		"page", "page_size", "order_by",
		"search", "search_fields", "preload",
		"fields",
	}
	for _, param := range specialParams {
		if key == param {
			return true
		}
	}
	return false
}
