package crud

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kruily/gofastcrud/core/crud/options"
	"github.com/kruily/gofastcrud/pkg/errors"
	"github.com/kruily/gofastcrud/pkg/validator"
)

// Create 创建实体
func (c *CrudController[T]) Create(ctx *gin.Context) (interface{}, error) {
	var entity T
	if err := ctx.ShouldBindJSON(&entity); err != nil {
		return nil, err
	}

	// 验证实体
	if err := validator.Validate(entity); err != nil {
		return nil, err
	}

	err := c.Repository.Create(ctx, &entity)
	if err != nil {
		return nil, err
	}

	return c.Responser.Success(entity), nil
}

// GetById 根据ID获取实体
func (c *CrudController[T]) GetById(ctx *gin.Context) (interface{}, error) {
	id := ctx.Param("id")
	if id == "" {
		return nil, errors.New(errors.ErrNotFound, "missing id parameter")
	}

	idInt, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, errors.New(errors.ErrNotFound, "invalid id format")
	}

	entity, err := c.Repository.FindById(ctx, uint(idInt))
	if err != nil {
		return nil, err
	}

	if entity == nil {
		return nil, errors.New(errors.ErrNotFound, "record not found")
	}

	return c.Responser.Success(entity), nil
}

// List 获取实体列表
func (c *CrudController[T]) List(ctx *gin.Context) (interface{}, error) {
	// 构建查询选项
	opts := c.buildQueryOptions(ctx)

	// 执行查询
	items, err := c.Repository.Find(ctx, &c.entity, opts)
	if err != nil {
		return nil, err
	}

	// 获取总数
	total, err := c.Repository.Count(ctx, &c.entity)
	if err != nil {
		return nil, err
	}

	return c.Responser.Pagenation(items, total, opts.Page, opts.PageSize), nil
}

// Update 更新实体
func (c *CrudController[T]) Update(ctx *gin.Context) (interface{}, error) {
	id := ctx.Param("id")
	if id == "" {
		return nil, errors.New(errors.ErrNotFound, "missing id parameter")
	}

	var entity T
	if err := ctx.ShouldBindJSON(&entity); err != nil {
		return nil, err
	}

	// 验证实体
	if err := validator.Validate(entity); err != nil {
		return nil, err
	}

	idInt, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, errors.New(errors.ErrNotFound, "invalid id format")
	}

	entity.SetID(uint(idInt))

	if err := c.Repository.Update(ctx, &entity); err != nil {
		return nil, err
	}

	return c.Responser.Success(entity), nil
}

// Delete 删除实体
func (c *CrudController[T]) Delete(ctx *gin.Context) (interface{}, error) {
	id := ctx.Param("id")
	if id == "" {
		return nil, errors.New(errors.ErrNotFound, "missing id parameter")
	}

	idInt, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, errors.New(errors.ErrNotFound, "invalid id format")
	}

	opts := options.NewDeleteOptions()
	if err := c.Repository.DeleteById(ctx, uint(idInt), opts); err != nil {
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
	var ids []uint
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
