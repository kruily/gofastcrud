package swagger

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/kruily/gofastcrud/core/crud"
	"github.com/kruily/gofastcrud/core/crud/types"
)

// Generator Swagger 文档生成器
type Generator struct {
	docs           map[string]*spec.Swagger
	processedTypes map[reflect.Type]*spec.Schema
}

// NewGenerator 创建生成器实例
func NewGenerator() *Generator {
	return &Generator{
		docs:           make(map[string]*spec.Swagger),
		processedTypes: make(map[reflect.Type]*spec.Schema),
	}
}

// RegisterEntityWithVersion 注册带版本的实体文档
func (g *Generator) RegisterEntityWithVersion(entityType reflect.Type, basePath string, routePath string, controller interface{}, version string) {
	// 处理指针类型
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}
	entityName := entityType.Name()
	paths := make(map[string]spec.PathItem)

	// 获取所有路由
	var allRoutes []types.APIRoute
	switch c := controller.(type) {
	case *crud.CrudController[crud.ICrudEntity]:
		allRoutes = c.GetRoutes()
	case interface{ GetRoutes() []types.APIRoute }:
		allRoutes = c.GetRoutes()
	}

	// 按路径分组路由
	routeGroups := make(map[string][]types.APIRoute)
	for _, route := range allRoutes {
		path := fmt.Sprintf("/%s%s", routePath, route.Path)
		re := regexp.MustCompile(`:\w+_id`)
		query := re.FindAllString(path, -1)
		for _, q := range query {
			if q != "" {
				rps := strings.TrimPrefix(q, ":")
				rps = "{" + rps + "}"
				path = strings.Replace(path, q, rps, 1)
			}
		}
		routeGroups[path] = append(routeGroups[path], route)
	}

	// 处理每个路径的所有方法
	for path, routes := range routeGroups {
		pathItem := spec.PathItem{}
		for _, route := range routes {
			operation := g.generateOperation(route, entityName)
			switch route.Method {
			case "GET":
				pathItem.Get = operation
			case "POST":
				pathItem.Post = operation
			case "PUT":
				pathItem.Put = operation
			case "DELETE":
				pathItem.Delete = operation
			}
		}
		paths[path] = pathItem
	}

	// 收集所有相关的模型定义
	definitions := make(spec.Definitions)
	g.processedTypes = make(map[reflect.Type]*spec.Schema) // 重置已处理类型的map
	definitions[entityName] = *g.generateSchema(entityType)

	// 收集请求和响应模型
	for _, routes := range routeGroups {
		for _, route := range routes {
			if route.Request != nil {
				reqType := reflect.TypeOf(route.Request)
				if reqType.Kind() == reflect.Ptr {
					reqType = reqType.Elem()
				}
				reqName := reqType.Name()
				if reqName != "" && reqName != entityName {
					g.processedTypes = make(map[reflect.Type]*spec.Schema) // 重置已处理类型的map
					definitions[reqName] = *g.generateSchema(reqType)
				}
			}
			if route.Response != nil {
				respType := reflect.TypeOf(route.Response)
				if respType.Kind() == reflect.Ptr {
					respType = respType.Elem()
				}
				respName := respType.Name()
				if respName != "" && respName != entityName {
					g.processedTypes = make(map[reflect.Type]*spec.Schema) // 重置已处理类型的map
					definitions[respName] = *g.generateSchema(respType)
				}
			}
		}
	}

	swagger := &spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Info: &spec.Info{
				InfoProps: spec.InfoProps{
					Title:       fmt.Sprintf("%s API", entityName),
					Description: fmt.Sprintf("API documentation for %s", entityName),
					Version:     version,
				},
			},
			BasePath:    basePath,
			Paths:       &spec.Paths{Paths: paths},
			Definitions: definitions,
			Tags: []spec.Tag{
				{
					TagProps: spec.TagProps{
						Name:        entityName,
						Description: fmt.Sprintf("Operations about %s", entityName),
					},
				},
			},
		},
	}

	g.docs[fmt.Sprintf("%s_%s", routePath, version)] = swagger
}

// GetSwagger 获取指定实体的 Swagger 文档
func (g *Generator) GetSwagger(entityPath string) *spec.Swagger {
	return g.docs[entityPath]
}

// GetAllSwagger 获取合并后的完整 Swagger 文档
func (g *Generator) GetAllSwagger() interface{} {
	versionDocs := make(map[string]*spec.Swagger)

	// 遍历所有文档，按版本分组
	for path, swagger := range g.docs {
		parts := strings.Split(path, "_")
		if len(parts) < 2 {
			continue
		}
		version := parts[len(parts)-1] // 获取版本号

		// 如果该版本的文档不存在，创建一个新的
		if _, exists := versionDocs[version]; !exists {
			versionDocs[version] = &spec.Swagger{
				SwaggerProps: spec.SwaggerProps{
					Swagger: "2.0",
					Info: &spec.Info{
						InfoProps: spec.InfoProps{
							Title:       fmt.Sprintf("Fast CRUD API (%s)", version),
							Description: fmt.Sprintf("Auto-generated API documentation for version %s", version),
							Version:     version,
						},
					},
					BasePath:    fmt.Sprintf("/api/%s", version),
					Schemes:     []string{"http"},
					Consumes:    []string{"application/json"},
					Produces:    []string{"application/json"},
					Paths:       &spec.Paths{Paths: make(map[string]spec.PathItem)},
					Definitions: make(spec.Definitions),
					Tags:        []spec.Tag{},
				},
			}
		}

		// 合并路径
		for path, item := range swagger.Paths.Paths {
			versionDocs[version].Paths.Paths[path] = item
		}

		// 合并定义
		for name, schema := range swagger.Definitions {
			if _, exists := versionDocs[version].Definitions[name]; !exists {
				versionDocs[version].Definitions[name] = schema
			}
		}

		// 合并标签（去重）
		if swagger.Tags != nil {
			tagMap := make(map[string]bool)
			for _, existingTag := range versionDocs[version].Tags {
				tagMap[existingTag.Name] = true
			}
			for _, tag := range swagger.Tags {
				if !tagMap[tag.Name] {
					versionDocs[version].Tags = append(versionDocs[version].Tags, tag)
				}
			}
		}
	}

	return versionDocs
}

// generateSchema 生成实体的 Schema
func (g *Generator) generateSchema(t reflect.Type) *spec.Schema {
	// 处理指针类型
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// 创建基础schema
	schema := &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:       []string{"object"},
			Properties: make(map[string]spec.Schema),
			Required:   []string{},
		},
	}

	// 处理字段
	if t.Kind() == reflect.Struct {
		// 检查是否已经处理过该类型
		if _, exists := g.processedTypes[t]; exists {
			return &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Ref: spec.MustCreateRef(fmt.Sprintf("#/definitions/%s", t.Name())),
				},
			}
		}
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
				for name, prop := range embeddedSchema.Properties {
					schema.Properties[name] = prop
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
		// 将schema加入到已处理map中，避免循环引用
		g.processedTypes[t] = schema
	} else if t.Kind() == reflect.Slice {
		elemSchema := g.generateSchema(t.Elem())
		schema.Type = []string{"array"}
		schema.Items = &spec.SchemaOrArray{
			Schema: elemSchema,
		}
	}

	return schema
}

// getFieldSchema 获取字段的 Schema
func (g *Generator) getFieldSchema(field reflect.StructField) spec.Schema {
	schema := spec.Schema{
		SchemaProps: spec.SchemaProps{},
	}

	// 处理字段类型
	switch field.Type.Kind() {
	case reflect.String:
		schema.Type = []string{"string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		schema.Type = []string{"integer"}
		if field.Type.Kind() == reflect.Int64 {
			schema.Format = "int64"
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema.Type = []string{"integer"}
		if field.Type.Kind() == reflect.Uint64 {
			schema.Format = "int64"
		}
	case reflect.Float32, reflect.Float64:
		schema.Type = []string{"number"}
		if field.Type.Kind() == reflect.Float64 {
			schema.Format = "double"
		}
	case reflect.Bool:
		schema.Type = []string{"boolean"}
	case reflect.Struct:
		if field.Type.String() == "time.Time" {
			schema.Type = []string{"string"}
			schema.Format = "date-time"
		} else {
			return *g.generateSchema(field.Type)
		}
	case reflect.Ptr:
		return *g.generateSchema(field.Type.Elem())
	case reflect.Slice:
		elemSchema := g.generateSchema(field.Type.Elem())
		schema.Type = []string{"array"}
		schema.Items = &spec.SchemaOrArray{
			Schema: elemSchema,
		}
	}

	// 添加描述
	if description := field.Tag.Get("description"); description != "" {
		schema.Description = description
	}

	// 添加示例
	if example := field.Tag.Get("example"); example != "" {
		schema.Example = example
	}

	return schema
}

// generateOperation 生成操作文档
func (g *Generator) generateOperation(route types.APIRoute, entityName string) *spec.Operation {
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

	operation := &spec.Operation{
		OperationProps: spec.OperationProps{
			Tags:        route.Tags,
			Summary:     route.Summary,
			Description: route.Description,
			ID:          operationId,
			Responses: &spec.Responses{
				ResponsesProps: spec.ResponsesProps{
					StatusCodeResponses: make(map[int]spec.Response),
				},
			},
		},
	}

	// 处理路径参数
	if re.MatchString(route.Path) {
		name := strings.Replace(route.Path, "/:", "", 1)
		operation.Parameters = append(operation.Parameters, spec.Parameter{
			ParamProps: spec.ParamProps{
				Name:        name,
				In:          "path",
				Description: "Entity ID",
				Required:    true,
			},
			SimpleSchema: spec.SimpleSchema{Type: route.PathType}, // 根据实际类型设置正确的类型
		})
	}

	// 处理查询参数
	if len(route.Parameters) > 0 {
		for _, param := range route.Parameters {
			parameter := spec.Parameter{
				ParamProps: spec.ParamProps{
					Name:        param.Name,
					In:          param.In,
					Description: param.Description,
					Required:    param.Required,
				},
				SimpleSchema: spec.SimpleSchema{
					Type:    param.Schema.Type,
					Format:  param.Schema.Format,
					Default: param.Schema.Default,
				},
			}
			operation.Parameters = append(operation.Parameters, parameter)
		}
	}

	// 添加请求体
	if route.Method == "POST" || route.Method == "PUT" {
		var schema *spec.Schema
		if route.Request != nil {
			schema = g.generateSchema(reflect.TypeOf(route.Request))
		} else {
			schema = &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Ref: spec.MustCreateRef(fmt.Sprintf("#/definitions/%s", entityName)),
				},
			}
		}
		operation.Parameters = append(operation.Parameters, spec.Parameter{
			ParamProps: spec.ParamProps{
				Name:        "body",
				In:          "body",
				Description: "Request body",
				Required:    true,
				Schema:      schema,
			},
		})
	}

	// 添加响应体
	if route.Response != nil {
		respSchema := g.generateSchema(reflect.TypeOf(route.Response))
		operation.Responses.StatusCodeResponses = map[int]spec.Response{
			200: {
				ResponseProps: spec.ResponseProps{
					Description: "Success",
					Schema:      respSchema,
				},
			},
		}
	}

	// 添加错误响应
	operation.Responses.StatusCodeResponses[400] = spec.Response{
		ResponseProps: spec.ResponseProps{
			Description: "Bad Request",
		},
	}
	operation.Responses.StatusCodeResponses[401] = spec.Response{
		ResponseProps: spec.ResponseProps{
			Description: "Unauthorized",
		},
	}
	operation.Responses.StatusCodeResponses[403] = spec.Response{
		ResponseProps: spec.ResponseProps{
			Description: "Forbidden",
		},
	}
	operation.Responses.StatusCodeResponses[404] = spec.Response{
		ResponseProps: spec.ResponseProps{
			Description: "Not Found",
		},
	}
	operation.Responses.StatusCodeResponses[500] = spec.Response{
		ResponseProps: spec.ResponseProps{
			Description: "Internal Server Error",
		},
	}

	return operation
}
