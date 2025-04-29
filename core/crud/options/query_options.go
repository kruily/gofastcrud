package options

import (
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
