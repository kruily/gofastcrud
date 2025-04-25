package crud

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/kruily/gofastcrud/core/crud/types"
)

// standardRoutes 标准路由
func (c *CrudController[T]) standardRoutes(cache bool, cacheTTL int) []types.APIRoute {
	entityName := strings.ToLower(c.entityName[:1]) + c.entityName[1:]
	idType := "integer"
	if _, ok := c.entity.GetID().(uuid.UUID); ok {
		idType = "string"
	}
	return []types.APIRoute{
		{
			Path:        "/:" + entityName + "_id",
			Method:      "GET",
			PathType:    idType,
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
			PathType:    idType,
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
			PathType:    idType,
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
			Response:    c.Responser.Success("批量创建成功"),
			Cache:       types.Cache{Enable: cache, Key: fmt.Sprintf("%s:%s", c.entityName, "batchCreate"), TTL: cacheTTL},
		},
		{
			Path:        "/batch",
			Method:      "PUT",
			Tags:        []string{c.entityName},
			Summary:     fmt.Sprintf("Batch Update %s", c.entityName),
			Description: fmt.Sprintf("Update multiple %s records", c.entityName),
			Handler:     c.BatchUpdate,
			Parameters:  c.modeParams(),
			Request:     []T{},
			Response:    c.Responser.Success("批量更新成功"),
			Cache:       types.Cache{Enable: cache, Key: fmt.Sprintf("%s:%s", c.entityName, "batchUpdate"), TTL: cacheTTL},
		},
		{
			Path:        "/batch",
			Method:      "DELETE",
			Tags:        []string{c.entityName},
			Summary:     fmt.Sprintf("Batch Delete %s", entityName),
			Description: fmt.Sprintf("Delete multiple %s records", entityName),
			Handler:     c.BatchDelete,
			Parameters:  c.modeParams(),
			Response:    c.Responser.Success("批量删除成功"),
			Cache:       types.Cache{Enable: cache, Key: fmt.Sprintf("%s:%s", c.entityName, "batchDelete"), TTL: cacheTTL},
		},
	}
}
