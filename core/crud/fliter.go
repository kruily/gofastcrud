package crud

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/kruily/gofastcrud/core/crud/types"
	"github.com/kruily/gofastcrud/errors"
)

// fliter tag 过滤条件 gt gte lt lte eq neq in nin like nlike between null all(全部) 以;分割

// QueryField 可查询字段 根据实体的fliter tag 过滤
// fliter tag 过滤条件 gt gte lt lte eq neq in nin like nlike between null all(全部) 以;分割
type QueryField struct {
	Field     string
	FilterTag string
}

// queryFields 从结构体中获取获取可查询字段
func queryFields(entity any) []QueryField {
	entityType := reflect.TypeOf(entity)
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}
	// 收集所有可查询字段
	fields := make([]QueryField, 0)
	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)
		// 跳过非导出字段和特殊字段
		if !field.IsExported() ||
			// field.Type.Kind() == reflect.Struct ||field.Type.Kind() == reflect.Ptr
			field.Type.Kind() == reflect.Slice ||
			field.Type.Kind() == reflect.Map {
			continue
		}
		if field.Anonymous && field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
			// 递归处理嵌套结构体
			subField := field.Type.Elem()
			subs := queryFields(reflect.New(subField).Interface().(any))
			fields = append(fields, subs...)
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}
		// 获取字段名（优先使用json tag）
		qf := QueryField{
			Field:     field.Name,
			FilterTag: field.Tag.Get("filter"),
		}
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				qf.Field = parts[0]
			}
		}
		fields = append(fields, qf)
	}
	return fields
}

// 过滤tag描述
var tagDescriptions = map[string]string{
	"gt":      "Greater than filter for %s, example: %s_gt=10",
	"gte":     "Greater than or equal filter for %s, example: %s_gte=10",
	"lt":      "Less than filter for %s, example: %s_lt=10",
	"lte":     "Less than or equal filter for %s, example: %s_lte=10",
	"eq":      "Equal filter for %s, example: %s_eq=10",
	"neq":     "Not equal filter for %s, example: %s_neq=10",
	"in":      "IN filter for %s (comma-separated values), example: %s_in=1,2,3",
	"nin":     "NOT IN filter for %s (comma-separated values), example: %s_nin=1,2,3",
	"like":    "LIKE filter for %s, example: %s_like=test",
	"nlike":   "NOT LIKE filter for %s, example: %s_nlike=test",
	"between": "BETWEEN filter for %s (comma-separated values), example: %s_between=1,10",
	"null":    "NULL filter for %s (true|false), example: %s_null=true",
}

// 生成模式查询参数
func generateModeQueryParams(queryParams []types.Parameter, queryFields []QueryField) ([]types.Parameter, error) {
	allFilterTags := []string{"gt", "gte", "lt", "lte", "eq", "neq", "in", "nin", "like", "nlike", "between", "null"}
	// 为每个字段添加过滤参数
	for _, field := range queryFields {
		tags := strings.Split(field.FilterTag, ",")
		// 如果没有指定tag，跳过
		if len(tags) == 0 || tags[0] == "" {
			continue
		}
		// 检查tag是否合法
		for _, tag := range tags {
			found := false
			for _, v := range allFilterTags {
				if v == tag {
					found = true
					break
				}
			}
			if !found {
				return nil, errors.New(errors.ErrInternal, fmt.Sprintf("filter tag %s is not valid", tag))
			}
		}
		// 如果 tag 为 all，添加所有tag
		if len(tags) == 1 && tags[0] == "all" {
			tags = allFilterTags
		}

		// 创建查询参数 openapi 描述
		for _, tag := range tags {
			queryParams = append(queryParams, types.Parameter{
				Name:        field.Field + "_" + tag,
				In:          "query",
				Description: fmt.Sprintf(tagDescriptions[tag], field.Field, field.Field),
				Schema:      types.Schema{Type: "string"},
			})
		}
	}

	return queryParams, nil
}
