package crud

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/kruily/gofastcrud/core/crud/types"
)

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
