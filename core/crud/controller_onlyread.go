package crud

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kruily/gofastcrud/core/crud/types"
	"github.com/kruily/gofastcrud/core/database"
	"github.com/kruily/gofastcrud/errors"
)

type OnlyReadController[T ICrudEntity] struct {
	*BlankController[T]
}

func NewOnlyReadController[T ICrudEntity](db *database.Database, entity T) *OnlyReadController[T] {
	controller := &OnlyReadController[T]{
		BlankController: NewBlankController(db, entity),
	}

	controller.routes = append(controller.routes, controller.standardRoutes(false, 0)...)

	return controller
}

// GetById 根据ID获取实体
func (c *OnlyReadController[T]) GetById(ctx *gin.Context) (interface{}, error) {
	// TODO 仍旧获取不到正确的ID
	id := ctx.Param(strings.ToLower(c.entityName) + "_id")
	if id == "" {
		return nil, errors.New(errors.ErrNotFound, "missing id parameter")
	}
	var idTID any

	// 处理 UUID 类型
	if idUUID, err := uuid.Parse(id); err == nil {
		idTID = idUUID // 直接赋值
	} else if idInt, err := strconv.ParseUint(id, 10, 64); err == nil {
		idTID = idInt // 直接赋值
	} else {
		return nil, errors.New(errors.ErrInvalidParam, "invalid id parameter")
	}

	entity, err := c.Repository.FindById(ctx, idTID)
	if err != nil {
		return nil, err
	}

	return c.Responser.Success(entity), nil
}

// List 获取实体列表
func (c *OnlyReadController[T]) List(ctx *gin.Context) (interface{}, error) {
	// 构建查询选项
	opts := c.BuildQueryOptions(ctx)

	// 执行查询
	items, err := c.Repository.Find(ctx, c.entity, opts)
	if err != nil {
		return nil, err
	}

	// 获取总数
	total, err := c.Repository.Count(ctx, c.entity)
	if err != nil {
		return nil, err
	}

	return c.Responser.Pagenation(items, total, opts.Page, opts.PageSize), nil
}

// standardRoutes 标准路由
func (c *OnlyReadController[T]) standardRoutes(cache bool, cacheTTL int) []*types.APIRoute {
	entityName := strings.ToLower(c.entityName[:1]) + c.entityName[1:]
	idType := "integer"
	if _, ok := c.entity.GetID().(uuid.UUID); ok {
		idType = "string"
	}
	return []*types.APIRoute{
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
	}
}
