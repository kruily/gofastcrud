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
	"github.com/kruily/gofastcrud/validator"
)

// CrudController 控制器实现
type CrudController[T ICrudEntity] struct {
	*BlankController[T]
}

// NewCrudController 创建控制器
func NewCrudController[T ICrudEntity](db *database.Database, entity T) *CrudController[T] {
	c := &CrudController[T]{
		BlankController: NewBlankController(db, entity),
	}
	c.routes = append(c.routes, c.standardRoutes(false, 0)...)

	// 自动配置预加载
	// c.configurePreloads()

	return c
}

// Create 创建实体
func (c *CrudController[T]) Create(ctx *gin.Context) (interface{}, error) {
	entity := NewModel[T]()
	if err := ctx.ShouldBindJSON(entity); err != nil {
		return nil, err
	}

	// 验证实体
	if err := validator.Validate(entity); err != nil {
		return nil, err
	}

	err := c.Repository.Create(ctx, entity)
	if err != nil {
		return nil, err
	}

	return c.Responser.Success(entity), nil
}

// GetById 根据ID获取实体
func (c *CrudController[T]) GetById(ctx *gin.Context) (interface{}, error) {
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
		idTID = id
	}

	entity, err := c.Repository.FindById(ctx, idTID)
	if err != nil {
		return nil, err
	}

	return c.Responser.Success(entity), nil
}

// List 获取实体列表
func (c *CrudController[T]) List(ctx *gin.Context) (interface{}, error) {
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

// Update 更新实体
func (c *CrudController[T]) Update(ctx *gin.Context) (interface{}, error) {
	id := ctx.Param(strings.ToLower(c.entityName) + "_id")
	if id == "" {
		return nil, errors.New(errors.ErrNotFound, "missing id parameter")
	}

	// 将请求体绑定到map
	updateFields := make(map[string]interface{})
	if err := ctx.ShouldBindJSON(&updateFields); err != nil {
		return nil, err
	}

	// 验证字段
	if err := validator.ValidateMap(updateFields, c.entity); err != nil {
		return nil, err
	}

	var idTID any

	// 处理 UUID 类型
	if idUUID, err := uuid.Parse(id); err == nil {
		idTID = idUUID // 直接赋值
	} else if idInt, err := strconv.ParseUint(id, 10, 64); err == nil {
		idTID = idInt // 直接赋值
	}else {
		idTID = id // 直接赋值
	}


	entity, err := c.Repository.FindById(ctx, idTID)
	if err != nil {
		return nil, errors.New(errors.ErrNotFound, "No record of this id was found")
	}

	// 更新指定字段
	if err := c.Repository.Update(ctx, entity, updateFields); err != nil {
		return nil, err
	}

	return c.Responser.Success(entity), nil
}

// Delete 删除实体
func (c *CrudController[T]) Delete(ctx *gin.Context) (interface{}, error) {
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
	}else {
		idTID = id // 直接赋值
	}


	// opts := options.NewDeleteOptions()
	if err := c.Repository.DeleteById(ctx, idTID); err != nil {
		return nil, err
	}

	return c.Responser.Success(nil), nil
}

// BatchCreate 批量创建实体
func (c *CrudController[T]) BatchCreate(ctx *gin.Context) (interface{}, error) {
	var entities []T
	if err := ctx.ShouldBindJSON(&entities); err != nil {
		return nil, err
	}

	// 验证每个实体
	for _, entity := range entities {
		if err := validator.Validate(entity); err != nil {
			return nil, err
		}
	}

	// 使用事务进行批量创建
	err := c.Repository.Transaction(ctx, func(tx IRepository[T]) error {
		return tx.BatchCreate(ctx, entities)
	})

	if err != nil {
		return nil, err
	}

	return c.Responser.Success(entities), nil
}

// BatchUpdate 批量更新实体
func (c *CrudController[T]) BatchUpdate(ctx *gin.Context) (interface{}, error) {
	var entities []T
	if err := ctx.ShouldBindJSON(&entities); err != nil {
		return nil, err
	}

	// 验证每个实体
	for _, entity := range entities {
		if err := validator.Validate(entity); err != nil {
			return nil, err
		}
	}

	// 使用事务进行批量更新
	err := c.Repository.Transaction(ctx, func(tx IRepository[T]) error {
		return tx.BatchUpdate(ctx, entities)
	})

	if err != nil {
		return nil, err
	}

	return c.Responser.Success(entities), nil
}

// BatchDelete 批量删除实体
func (c *CrudController[T]) BatchDelete(ctx *gin.Context) (interface{}, error) {
	var ids []any
	if err := ctx.ShouldBindJSON(&ids); err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return nil, errors.New(errors.ErrInvalidParam, "no ids provided")
	}

	// 使用事务进行批量删除
	err := c.Repository.Transaction(ctx, func(tx IRepository[T]) error {
		return tx.BatchDelete(ctx, ids)
	})

	if err != nil {
		return nil, err
	}

	return c.Responser.Success(nil), nil
}

// standardRoutes 标准路由
func (c *CrudController[T]) standardRoutes(cache bool, cacheTTL int) []*types.APIRoute {
	entityName := strings.ToLower(c.entityName[:1]) + c.entityName[1:]
	idType := "string"
	if _, ok := c.entity.GetID().(uint); ok {
		idType = "integer"
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
			Parameters:  ModeParams(c),
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
			Parameters:  ModeParams(c),
			Response:    c.Responser.Success("批量删除成功"),
			Cache:       types.Cache{Enable: cache, Key: fmt.Sprintf("%s:%s", c.entityName, "batchDelete"), TTL: cacheTTL},
		},
	}
}
