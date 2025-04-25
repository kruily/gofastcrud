package crud

import (
	"strings"

	"github.com/kruily/gofastcrud/core/crud/types"
)

// queryParams 获取查询参数
func (c *CrudController[T]) queryParams() []types.Parameter {
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
	queryParams = append(queryParams, c.modeParams()...)
	return queryParams
}

// deleteParams 获取删除参数
func (c *CrudController[T]) modeParams() []types.Parameter {
	// 获取所有可查询字段
	queryFields := queryFields(c.entity)
	params := []types.Parameter{}
	params, err := generateModeQueryParams(params, queryFields)
	if err != nil {
		panic(err)
	}
	return params
}
