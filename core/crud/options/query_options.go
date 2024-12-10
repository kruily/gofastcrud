package options

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// QueryOptions 查询选项
type QueryOptions struct {
	// 分页
	Page     int
	PageSize int
	// 排序
	OrderBy []string
	// 查询条件
	Where map[string]interface{}
	// 预加载关系
	Preload []string
	// 选择特定字段
	Select []string
	// 搜索关键词
	Search string
	// 搜索字段
	SearchFields []string
	// 过滤条件
	Filter map[string]interface{}
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

// NewQueryOptions 创建查询选项
func NewQueryOptions(opts ...func(*QueryOptions)) *QueryOptions {
	// 创建默认选项
	options := &QueryOptions{
		Where:  make(map[string]interface{}),
		Filter: make(map[string]interface{}),
	}
	// 应用开发者传入的选项
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// WithPage 设置分页
func WithPage(page int) func(*QueryOptions) {
	return func(q *QueryOptions) {
		q.Page = page
	}
}

// WithPageSize 设置页面大小
func WithPageSize(pageSize int) func(*QueryOptions) {
	return func(q *QueryOptions) {
		q.PageSize = pageSize
	}
}

// WithOrderBy 设置排序
func WithOrderBy(orderBy string) func(*QueryOptions) {
	return func(q *QueryOptions) {
		q.OrderBy = []string{orderBy}
	}
}

// WithSearch 设置搜索
func WithSearch(search string) func(*QueryOptions) {
	return func(q *QueryOptions) {
		q.Search = search
	}
}

// WithSearchFields 设置搜索字段
func WithSearchFields(searchFields []string) func(*QueryOptions) {
	return func(q *QueryOptions) {
		q.SearchFields = searchFields
	}
}

// WithPreload 设置预加载
func WithPreload(preload []string) func(*QueryOptions) {
	return func(q *QueryOptions) {
		q.Preload = preload
	}
}

// WithSelect 设置选择特定字段
func WithSelect(selects []string) func(*QueryOptions) {
	return func(q *QueryOptions) {
		q.Select = selects
	}
}

// WithFilter 设置过滤条件
func WithFilter(filter map[string]interface{}) func(*QueryOptions) {
	return func(q *QueryOptions) {
		q.Filter = filter
	}
}

// BuildQueryOptions 构建查询选项
func (q *QueryOptions) BuildQueryOptions(ctx *gin.Context) {
	// 处理搜索
	if search := ctx.Query("search"); search != "" {
		q.Search = search
		q.SearchFields = strings.Split(ctx.DefaultQuery("search_fields", "id"), ",")
	}

	// 处理过滤条件
	q.buildFilterOptions(ctx)

	// 处理预加载关系
	if preload := ctx.Query("preload"); preload != "" {
		q.Preload = strings.Split(preload, ",")
	}

	// 处理字段选择
	if fields := ctx.Query("fields"); fields != "" {
		q.Select = strings.Split(fields, ",")
	}
}

// buildFilterOptions 构建过滤选项
func (q *QueryOptions) buildFilterOptions(ctx *gin.Context) {
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
				q.Where[field+" > ?"] = values[0]
			} else if strings.HasSuffix(key, "_lt") {
				q.Where[field+" < ?"] = values[0]
			} else if strings.HasSuffix(key, "_gte") {
				q.Where[field+" >= ?"] = values[0]
			} else if strings.HasSuffix(key, "_lte") {
				q.Where[field+" <= ?"] = values[0]
			}
			continue
		}

		// 处理IN查询
		if strings.HasSuffix(key, "_in") {
			field := strings.TrimSuffix(key, "_in")
			q.Where[field+" IN ?"] = strings.Split(values[0], ",")
			continue
		}

		// 处理NULL查询
		if strings.HasSuffix(key, "_null") {
			field := strings.TrimSuffix(key, "_null")
			if values[0] == "true" {
				q.Where[field+" IS NULL"] = nil
			} else {
				q.Where[field+" IS NOT NULL"] = nil
			}
			continue
		}

		// 处理模糊查询
		if strings.HasSuffix(key, "_like") {
			field := strings.TrimSuffix(key, "_like")
			q.Where[field+" LIKE ?"] = "%" + values[0] + "%"
			continue
		}

		// 处理普通相等查询
		if !isSpecialParam(key) {
			q.Filter[key] = values[0]
		}
	}
}

// applyQueryOptions 应用查询选项
func (q *QueryOptions) ApplyQueryOptions(db *gorm.DB) *gorm.DB {
	// 应用搜索
	if q.Search != "" && len(q.SearchFields) > 0 {
		for _, field := range q.SearchFields {
			db = db.Or(field+" LIKE ?", "%"+q.Search+"%")
		}
	}

	// 应用过滤条件
	if len(q.Filter) > 0 {
		for key, value := range q.Filter {
			db = db.Where(key, value)
		}
	}

	// 应用查询条件
	if len(q.Where) > 0 {
		for key, value := range q.Where {
			db = db.Where(key, value)
		}
	}

	// 应用排序
	if len(q.OrderBy) > 0 {
		for _, order := range q.OrderBy {
			db = db.Order(order)
		}
	}

	// 应用预加载
	if len(q.Preload) > 0 {
		for _, preload := range q.Preload {
			db = db.Preload(preload)
		}
	}

	// 应用选择特定字段
	if len(q.Select) > 0 {
		db = db.Select(q.Select)
	}

	// 应用分页
	if q.Page > 0 && q.PageSize > 0 {
		offset := (q.Page - 1) * q.PageSize
		db = db.Offset(offset).Limit(q.PageSize)
	}

	return db
}
