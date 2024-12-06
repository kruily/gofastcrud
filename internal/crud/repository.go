package crud

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// IRepository 通用仓储接口
type IRepository[T ICrudEntity] interface {
	// Create 操作
	Create(ctx context.Context, entity *T) error
	CreateInBatches(ctx context.Context, entities []*T, batchSize int) error

	// Read 操作
	FindById(ctx context.Context, id interface{}) (*T, error)
	FindOne(ctx context.Context, query *T, opts ...QueryOptions) (*T, error)
	Find(ctx context.Context, query *T, opts ...QueryOptions) ([]T, error)
	Count(ctx context.Context, query *T) (int64, error)

	// Update 操作
	Update(ctx context.Context, entity *T) error
	UpdateFields(ctx context.Context, id interface{}, fields map[string]interface{}) error

	// Delete 操作
	Delete(ctx context.Context, query *T, opts ...DeleteOptions) error
	DeleteById(ctx context.Context, id interface{}, opts ...DeleteOptions) error

	// 事务相关
	Transaction(ctx context.Context, fn func(txRepo IRepository[T]) error) error
}

// Repository 通用仓储实现
type Repository[T ICrudEntity] struct {
	db     *gorm.DB
	entity T
}

// NewRepository 创建仓储实例
func NewRepository[T ICrudEntity](db *gorm.DB, entity T) *Repository[T] {
	return &Repository[T]{
		db:     db,
		entity: entity,
	}
}

// withContext 添加上下文
func (r *Repository[T]) withContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Table(r.entity.Table())
}

// applyQueryOptions 应用查询选项
func (r *Repository[T]) applyQueryOptions(db *gorm.DB, opts ...QueryOptions) *gorm.DB {
	if len(opts) == 0 {
		return db
	}
	opt := opts[0]

	// 应用分页
	if opt.Page > 0 && opt.PageSize > 0 {
		offset := (opt.Page - 1) * opt.PageSize
		db = db.Offset(offset).Limit(opt.PageSize)
	}

	// 应用排序
	for _, order := range opt.OrderBy {
		db = db.Order(order)
	}

	// 应用查询条件
	if len(opt.Where) > 0 {
		db = db.Where(opt.Where)
	}

	// 应用预加载
	for _, preload := range opt.Preload {
		db = db.Preload(preload)
	}

	// 应用字段选择
	if len(opt.Select) > 0 {
		db = db.Select(opt.Select)
	}

	return db
}

// Create 实现
func (r *Repository[T]) Create(ctx context.Context, entity *T) error {
	return r.withContext(ctx).Create(entity).Error
}

// CreateInBatches 实现
func (r *Repository[T]) CreateInBatches(ctx context.Context, entities []*T, batchSize int) error {
	return r.withContext(ctx).CreateInBatches(entities, batchSize).Error
}

// FindById 实现
func (r *Repository[T]) FindById(ctx context.Context, id interface{}) (*T, error) {
	var entity T
	err := r.withContext(ctx).First(&entity, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// Transaction 实现
func (r *Repository[T]) Transaction(ctx context.Context, fn func(txRepo IRepository[T]) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := NewRepository[T](tx, r.entity)
		return fn(txRepo)
	})
}

// Count 实现
func (r *Repository[T]) Count(ctx context.Context, query *T) (int64, error) {
	var count int64
	err := r.db.Model(query).Count(&count).Error
	return count, err
}

// Delete 实现
func (r *Repository[T]) Delete(ctx context.Context, query *T, opts ...DeleteOptions) error {
	db := r.withContext(ctx)
	if len(opts) > 0 && opts[0].Force {
		db = db.Unscoped()
	}
	return db.Delete(query).Error
}

// DeleteById 实现
func (r *Repository[T]) DeleteById(ctx context.Context, id interface{}, opts ...DeleteOptions) error {
	var entity T
	db := r.withContext(ctx)
	if len(opts) > 0 && opts[0].Force {
		db = db.Unscoped()
	}
	return db.Delete(&entity, id).Error
}

// Find 实现
func (r *Repository[T]) Find(ctx context.Context, query *T, opts ...QueryOptions) ([]T, error) {
	var entities []T
	db := r.applyQueryOptions(r.withContext(ctx), opts...)
	err := db.Find(&entities).Error
	return entities, err
}

// FindOne 实现
func (r *Repository[T]) FindOne(ctx context.Context, query *T, opts ...QueryOptions) (*T, error) {
	var entity T
	err := r.applyQueryOptions(r.withContext(ctx), opts...).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// Update 实现
func (r *Repository[T]) Update(ctx context.Context, entity *T) error {
	return r.withContext(ctx).Updates(entity).Error
}

// UpdateFields 实现
func (r *Repository[T]) UpdateFields(ctx context.Context, id interface{}, fields map[string]interface{}) error {
	return r.withContext(ctx).Model(new(T)).Where("id = ?", id).Updates(fields).Error
}
