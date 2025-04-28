package openapi

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/kruily/gofastcrud/core/crud/types"
)

// generateSchema 生成实体的 Schema
func (g *GeneratorV3) generateSchema(t reflect.Type) *openapi3.SchemaRef {
	// 处理指针类型
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// 检查是否已经处理过该类型
	if _, exists := g.processedTypes[t]; exists {
		return &openapi3.SchemaRef{
			Ref: fmt.Sprintf("#/components/schemas/%s", t.Name()),
		}
	}

	// 创建基础schema
	schema := &openapi3.Schema{
		Type:       &openapi3.Types{"object"},
		Properties: make(openapi3.Schemas),
		Required:   []string{},
	}

	// 将schema加入到已处理map中，避免循环引用
	g.processedTypes[t] = schema

	// 处理字段
	if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			// 跳过非导出字段
			if !field.IsExported() {
				continue
			}
			// 处理嵌入字段
			if field.Anonymous {
				if field.Type.Kind() == reflect.Ptr {
					field.Type = field.Type.Elem()
				}
				if field.Type.Name() == "BaseEntity" {
					continue
				}
				embeddedSchema := g.generateSchema(field.Type)
				if embeddedSchema.Value != nil {
					for name, prop := range embeddedSchema.Value.Properties {
						schema.Properties[name] = prop
					}
				}
				continue
			}
			// 处理 json 标签
			jsonTag := field.Tag.Get("json")
			if jsonTag == "-" || jsonTag == ",inline" {
				continue
			}
			name := field.Name
			if jsonTag != "" {
				parts := strings.Split(jsonTag, ",")
				if parts[0] != "" {
					name = parts[0]
				}
			}
			// 生成字段的 schema
			fieldSchema := g.getFieldSchema(field)
			schema.Properties[name] = fieldSchema
			// 处理必填字段
			if required := field.Tag.Get("binding"); required == "required" {
				schema.Required = append(schema.Required, name)
			}
		}
	}
	return &openapi3.SchemaRef{
		Value: schema,
	}
}

// getFieldSchema 获取字段的 Schema
func (g *GeneratorV3) getFieldSchema(field reflect.StructField) *openapi3.SchemaRef {
	schema := &openapi3.Schema{}

	// 处理字段类型
	switch field.Type.Kind() {
	case reflect.String:
		schema.Type = &openapi3.Types{"string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		schema.Type = &openapi3.Types{"integer"}
		if field.Type.Kind() == reflect.Int64 {
			schema.Format = "int64"
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema.Type = &openapi3.Types{"integer"}
		if field.Type.Kind() == reflect.Uint64 {
			schema.Format = "int64"
		}
	case reflect.Float32, reflect.Float64:
		schema.Type = &openapi3.Types{"number"}
		if field.Type.Kind() == reflect.Float64 {
			schema.Format = "double"
		}
	case reflect.Bool:
		schema.Type = &openapi3.Types{"boolean"}
	case reflect.Struct:
		if field.Type.String() == "time.Time" {
			schema.Type = &openapi3.Types{"string"}
			schema.Format = "date-time"
		} else {
			return g.generateSchema(field.Type)
		}
	case reflect.Ptr:
		return g.generateSchema(field.Type.Elem())
	case reflect.Slice:
		schema.Type = &openapi3.Types{"array"}
		schema.Items = g.generateSchema(field.Type.Elem())
	}

	// 添加描述
	if description := field.Tag.Get("description"); description != "" {
		schema.Description = description
	}

	// 添加示例
	if example := field.Tag.Get("example"); example != "" {
		schema.Example = example
	}

	return &openapi3.SchemaRef{Value: schema}
}

// generateOperation 生成操作文档
func (g *GeneratorV3) generateOperation(route *types.APIRoute, entityName string) *openapi3.Operation {
	// 生成 operationId
	operationId := ""
	if len(strings.Split(route.Path, "/")) > 1 {
		parts := strings.Split(route.Path, "/")
		var pathParts []string
		for _, part := range parts {
			if part != "" && !strings.HasPrefix(part, ":") {
				// 首字母大写
				pathParts = append(pathParts, strings.ToUpper(part[0:1])+part[1:])
			}
		}
		operationId = fmt.Sprintf("%s%s%s",
			strings.ToLower(route.Method),
			strings.Replace(entityName, " ", "", -1),
			strings.Join(pathParts, ""),
		)
	} else {
		operationId = fmt.Sprintf("%s%s",
			strings.ToLower(route.Method),
			strings.Replace(entityName, " ", "", -1),
		)
	}
	re := regexp.MustCompile(`:\w+_id`)
	if re.MatchString(route.Path) {
		operationId += "ById"
	}

	operation := &openapi3.Operation{
		Tags:        route.Tags,
		Summary:     route.Summary,
		Description: route.Description,
		OperationID: operationId,
		// Responses:   make(map[string]openapi3.Responses),
		Responses: &openapi3.Responses{},
	}

	// 处理路径参数
	pathParams := regexp.MustCompile(`/:([^/]+)`).FindAllStringSubmatch(route.Path, -1)
	for _, param := range pathParams {
		paramName := param[1]
		operation.Parameters = append(operation.Parameters, &openapi3.ParameterRef{
			Value: &openapi3.Parameter{
				Name:        paramName,
				In:          openapi3.ParameterInPath,
				Description: fmt.Sprintf("%s parameter", strings.Title(paramName)),
				Required:    true,
				Schema: &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: &openapi3.Types{route.PathType},
					},
				},
			},
		})
	}

	// 处理查询参数
	if len(route.Parameters) > 0 {
		for _, param := range route.Parameters {
			operation.Parameters = append(operation.Parameters, &openapi3.ParameterRef{
				Value: &openapi3.Parameter{
					Name:        param.Name,
					In:          param.In,
					Description: param.Description,
					Required:    param.Required,
					Schema: &openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:    &openapi3.Types{param.Schema.Type},
							Format:  param.Schema.Format,
							Default: param.Schema.Default,
						},
					},
				},
			})
		}
	}

	// 添加请求体
	if route.Method == "POST" || route.Method == "PUT" {
		var schema *openapi3.SchemaRef
		if route.Request != nil {
			t := reflect.TypeOf(route.Request)
			schemaValue := g.generateSchema(t)
			schema = schemaValue
		} else {
			schema = &openapi3.SchemaRef{
				Ref: fmt.Sprintf("#/components/schemas/%s", entityName),
			}
		}

		operation.RequestBody = &openapi3.RequestBodyRef{
			Value: &openapi3.RequestBody{
				Description: "Request body",
				Required:    true,
				Content: openapi3.Content{
					"application/json": &openapi3.MediaType{
						Schema: schema,
					},
				},
			},
		}
	}

	// 添加响应体
	if route.Response != nil {
		success := "Success"
		respSchema := g.generateSchema(reflect.TypeOf(route.Response))
		operation.Responses.Set("200", &openapi3.ResponseRef{
			Value: &openapi3.Response{
				Description: &success,
				Content: openapi3.Content{
					"application/json": &openapi3.MediaType{
						Schema: respSchema,
					},
				},
			},
		})
	}

	// 添加错误响应
	errorSchema := &openapi3.Schema{
		Type: &openapi3.Types{"object"},
		Properties: openapi3.Schemas{
			"code":    &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}}},
			"message": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
		},
	}

	errorResponse := &openapi3.Response{
		Content: openapi3.Content{
			"application/json": &openapi3.MediaType{
				Schema: &openapi3.SchemaRef{Value: errorSchema},
			},
		},
	}

	bad := "Bad Request"
	unau := "Unauthorized"
	notf := "Not Found"
	inte := "Internal Server Error"

	// 添加标准错误响应
	operation.Responses.Set("400", &openapi3.ResponseRef{
		Value: &openapi3.Response{Description: &bad, Content: errorResponse.Content}},
	)

	operation.Responses.Set("401", &openapi3.ResponseRef{
		Value: &openapi3.Response{Description: &unau, Content: errorResponse.Content}},
	)

	operation.Responses.Set("404", &openapi3.ResponseRef{
		Value: &openapi3.Response{Description: &notf, Content: errorResponse.Content}},
	)

	operation.Responses.Set("500", &openapi3.ResponseRef{
		Value: &openapi3.Response{Description: &inte, Content: errorResponse.Content}},
	)
	return operation
}
