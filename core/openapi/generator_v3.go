package openapi

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/kruily/gofastcrud/core/crud"
	"github.com/kruily/gofastcrud/core/crud/types"
)

// GeneratorV3 OpenAPI 3.0 文档生成器
type GeneratorV3 struct {
	docs           map[string]*openapi3.T
	processedTypes map[reflect.Type]*openapi3.Schema
}

// NewGeneratorV3 创建生成器实例
func NewGeneratorV3() *GeneratorV3 {
	return &GeneratorV3{
		docs:           make(map[string]*openapi3.T),
		processedTypes: make(map[reflect.Type]*openapi3.Schema),
	}
}

// RegisterEntityWithVersion 注册带版本的实体文档
func (g *GeneratorV3) RegisterEntityWithVersion(entityType reflect.Type, basePath string, routePath string, controller interface{}, version string) {
	// 处理指针类型
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}
	entityName := entityType.Name()
	paths := openapi3.Paths{}

	// 获取所有路由
	var allRoutes []*types.APIRoute
	switch c := controller.(type) {
	case *crud.CrudController[crud.ICrudEntity]:
		allRoutes = c.GetRoutes()
	case interface{ GetRoutes() []*types.APIRoute }:
		allRoutes = c.GetRoutes()
	}

	// 按路径分组路由
	routeGroups := make(map[string][]*types.APIRoute)
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
		pathItem := &openapi3.PathItem{}
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
		// paths[path] = pathItem
		paths.Set(path, pathItem)
	}

	// 创建OpenAPI文档
	openapi := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:       fmt.Sprintf("%s API", entityName),
			Description: fmt.Sprintf("API documentation for %s", entityName),
			Version:     version,
		},
		Paths: &paths,
		Components: &openapi3.Components{
			Schemas: make(openapi3.Schemas),
		},
	}

	// 收集所有相关的模型定义
	g.processedTypes = make(map[reflect.Type]*openapi3.Schema) // 重置已处理类型的map
	openapi.Components.Schemas[entityName] = g.generateSchema(entityType)

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
					g.processedTypes = make(map[reflect.Type]*openapi3.Schema) // 重置已处理类型的map
					openapi.Components.Schemas[reqName] = g.generateSchema(reqType)
				}
			}
			if route.Response != nil {
				respType := reflect.TypeOf(route.Response)
				if respType.Kind() == reflect.Ptr {
					respType = respType.Elem()
				}
				respName := respType.Name()
				if respName != "" && respName != entityName {
					g.processedTypes = make(map[reflect.Type]*openapi3.Schema) // 重置已处理类型的map
					openapi.Components.Schemas[respName] = g.generateSchema(respType)
				}
			}
		}
	}

	// 添加标签
	openapi.Tags = openapi3.Tags{{
		Name:        entityName,
		Description: fmt.Sprintf("Operations about %s", entityName),
	}}

	g.docs[fmt.Sprintf("%s_%s", routePath, version)] = openapi
}

// GetSwagger 获取指定实体的 OpenAPI 文档
func (g *GeneratorV3) GetSwagger(entityPath string) *openapi3.T {
	return g.docs[entityPath]
}

// GetAllSwagger 获取合并后的完整 OpenAPI 文档
func (g *GeneratorV3) GetAllSwagger() interface{} {
	versionDocs := make(map[string]*openapi3.T)

	// 遍历所有文档，按版本分组
	for path, openapi := range g.docs {
		parts := strings.Split(path, "_")
		if len(parts) < 2 {
			continue
		}
		version := parts[len(parts)-1] // 获取版本号

		// 如果该版本的文档不存在，创建一个新的
		if _, exists := versionDocs[version]; !exists {
			versionDocs[version] = &openapi3.T{
				OpenAPI: "3.0.0",
				Info: &openapi3.Info{
					Title:       fmt.Sprintf("Fast CRUD API (%s)", version),
					Description: fmt.Sprintf("Auto-generated API documentation for version %s", version),
					Version:     version,
				},
				Servers: openapi3.Servers{{
					URL:         fmt.Sprintf("/api/%s", version),
					Description: fmt.Sprintf("Version %s API", version),
				}},
				Paths: &openapi3.Paths{},
				Components: &openapi3.Components{
					Schemas: make(openapi3.Schemas),
				},
			}
		}

		// 合并路径
		for path, item := range openapi.Paths.Map() {
			// versionDocs[version].Paths[path] = item
			versionDocs[version].Paths.Set(path, item)
		}

		// 合并定义
		for name, schema := range openapi.Components.Schemas {
			if _, exists := versionDocs[version].Components.Schemas[name]; !exists {
				versionDocs[version].Components.Schemas[name] = schema
			}
		}

		// 合并标签（去重）
		if openapi.Tags != nil {
			tagMap := make(map[string]bool)
			for _, existingTag := range versionDocs[version].Tags {
				tagMap[existingTag.Name] = true
			}
			for _, tag := range openapi.Tags {
				if !tagMap[tag.Name] {
					versionDocs[version].Tags = append(versionDocs[version].Tags, tag)
				}
			}
		}
	}

	return versionDocs
}
