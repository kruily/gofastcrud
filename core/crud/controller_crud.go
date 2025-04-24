package crud

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kruily/gofastcrud/config"
	"github.com/kruily/gofastcrud/core/crud/options"
	"github.com/kruily/gofastcrud/errors"
	"github.com/kruily/gofastcrud/validator"
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
	id := ctx.Param(c.entityName + "_id")
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
	id := ctx.Param(c.entityName + "_id")
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

	entity.SetID(idInt)

	if err := c.Repository.Update(ctx, &entity); err != nil {
		return nil, err
	}

	return c.Responser.Success(entity), nil
}

// Delete 删除实体
func (c *CrudController[T]) Delete(ctx *gin.Context) (interface{}, error) {
	id := ctx.Param(c.entityName + "_id")
	if id == "" {
		return nil, errors.New(errors.ErrNotFound, "missing id parameter")
	}

	idInt, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, errors.New(errors.ErrNotFound, "invalid id format")
	}

	opts := options.NewDeleteOptions()
	if err := c.Repository.DeleteById(ctx, idInt, opts); err != nil {
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

// buildQueryOptions 构建查询选项
func (c *CrudController[T]) buildQueryOptions(ctx *gin.Context) *options.QueryOptions {
	// 获取基础分页参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", strconv.Itoa(config.CONFIG_MANAGER.GetConfig().Pagenation.DefaultPageSize)))

	// 限制页面大小
	if pageSize > config.CONFIG_MANAGER.GetConfig().Pagenation.MaxPageSize {
		pageSize = config.CONFIG_MANAGER.GetConfig().Pagenation.MaxPageSize
	}

	// 构建查询选项
	opts := options.NewQueryOptions(
		options.WithPage(page),
		options.WithPageSize(pageSize),
		options.WithOrderBy(ctx.DefaultQuery("order_by", "id desc")),
	)

	// 处理搜索
	if search := ctx.Query("search"); search != "" {
		opts.Search = search
		opts.SearchFields = strings.Split(ctx.DefaultQuery("search_fields", "id"), ",")
	}

	// 处理过滤条件
	c.buildFilterOptions(ctx, opts)

	// 处理预加载关系
	if preload := ctx.Query("preload"); preload != "" {
		opts.Preload = strings.Split(preload, ",")
	}

	// 处理字段选择
	if fields := ctx.Query("fields"); fields != "" {
		opts.Select = strings.Split(fields, ",")
	}

	return opts
}

// buildFilterOptions 构建过滤选项
func (c *CrudController[T]) buildFilterOptions(ctx *gin.Context, opts *options.QueryOptions) {
	querys := ctx.Request.URL.Query()
	for key, values := range querys {
		// 跳过特殊参数
		if isSpecialParam(key) {
			continue
		}
		// 处理范围查询
		if strings.HasSuffix(key, "_gt") || strings.HasSuffix(key, "_lt") ||
			strings.HasSuffix(key, "_gte") || strings.HasSuffix(key, "_lte") {
			field := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(
				strings.TrimSuffix(key, "_gt"), "_lt"), "_gte"), "_lte")

			if strings.HasSuffix(key, "_gt") {
				opts.Where[field+" > ?"] = values[0]
			} else if strings.HasSuffix(key, "_lt") {
				opts.Where[field+" < ?"] = values[0]
			} else if strings.HasSuffix(key, "_gte") {
				opts.Where[field+" >= ?"] = values[0]
			} else if strings.HasSuffix(key, "_lte") {
				opts.Where[field+" <= ?"] = values[0]
			}
			continue
		}

		// 处理IN查询
		if strings.HasSuffix(key, "_in") {
			field := strings.TrimSuffix(key, "_in")
			opts.Where[field+" IN ?"] = strings.Split(values[0], ",")
			continue
		}

		// 处理NULL查询
		if strings.HasSuffix(key, "_null") {
			field := strings.TrimSuffix(key, "_null")
			if values[0] == "true" {
				opts.Where[field+" IS NULL"] = nil
			} else {
				opts.Where[field+" IS NOT NULL"] = nil
			}
			continue
		}

		// 处理模糊查询
		if strings.HasSuffix(key, "_like") {
			field := strings.TrimSuffix(key, "_like")
			opts.Where[field+" LIKE ?"] = "%" + values[0] + "%"
			continue
		}

		// 处理普通相等查询
		if !isSpecialParam(key) {
			opts.Filter[key] = values[0]
		}
	}
}

// isSpecialParam 检查是否为特殊参数
func isSpecialParam(key string) bool {
	specialParams := []string{
		"page", "page_size", "order_by",
		"search", "search_fields", "preload",
		"fields",
	}
	for _, param := range specialParams {
		if key == param {
			return true
		}
	}
	return false
}
