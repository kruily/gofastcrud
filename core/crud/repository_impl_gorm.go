package crud

import (
	"context"
	"reflect"

	"github.com/kruily/gofastcrud/core/crud/options"
	"gorm.io/gorm"
)

// gormRepository gorm仓储实现
type gormRepository[T ICrudEntity] struct {
	db         *gorm.DB
	entityType reflect.Type
	preloads   []string
}

// newGormRepository 创建gorm仓储实例
func newGormRepository[T ICrudEntity](db *gorm.DB, entity T) *gormRepository[T] {
	entityType := reflect.TypeOf(entity)
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}
	return &gormRepository[T]{
		db:         db,
		entityType: entityType,
		preloads:   make([]string, 0),
	}
}

// AddQueryHook 添加查询钩子
// func (r *gormRepository[T]) AddQueryHook(hook QueryHook) IRepository[T] {
// 	// 创建新的会话以避免污染原有查询
// 	db := r.db.Session(&gorm.Session{})
// 	// 注册回调
// 	db.Callback().Query().Before("gorm:query").Register("my_hook:before", hook.BeforeQuery)
// 	db.Callback().Query().After("gorm:query").Register("my_hook:after", hook.AfterQuery)
// 	r.db = db
// 	return r
// }

// applyPreloads 应用预加载
func (r *gormRepository[T]) applyPreloads(db *gorm.DB) *gorm.DB {
	for _, preload := range r.preloads {
		db = db.Preload(preload)
	}
	return db
}

// FindOne 查询单个实体
func (r *gormRepository[T]) FindOne(ctx context.Context, query interface{}, args ...interface{}) (T, error) {
	// var entity T
	entity := NewModel[T]()
	db := r.applyPreloads(r.db.WithContext(ctx))
	err := db.Where(query, args...).First(&entity).Error
	if err != nil {
		return entity, err
	}
	return entity, nil
}

// Find 查询实体列表
func (r *gormRepository[T]) Find(ctx context.Context, entity T, opts *options.QueryOptions) ([]T, error) {
	var entities []T
	db := r.applyPreloads(r.db.WithContext(ctx).Model(&entity))

	// 应用查询选项
	db = opts.ApplyQueryOptions(db)

	err := db.Find(&entities).Error
	return entities, err
}

// FindAll 查询所有符合条件的记录
func (r *gormRepository[T]) FindAll(ctx context.Context, query interface{}, args ...interface{}) ([]T, error) {
	var entities []T
	db := r.applyPreloads(r.db.WithContext(ctx))
	err := db.Where(query, args...).Find(&entities).Error
	return entities, err
}

// FindById 根据ID查询
func (r *gormRepository[T]) FindById(ctx context.Context, id any) (T, error) {
	// var entity T
	entity := NewModel[T]()
	db := r.applyPreloads(r.db.WithContext(ctx))
	err := db.First(&entity, id).Error
	if err != nil {
		return entity, err
	}
	return entity, nil
}

// 实现所有接口方法...
func (r *gormRepository[T]) Create(ctx context.Context, entity T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// BatchCreate 批量创建
func (r *gormRepository[T]) BatchCreate(ctx context.Context, entities []T, opts ...*options.BatchOptions) error {
	batchSize := 100
	if len(opts) > 0 && opts[0].BatchSize > 0 {
		batchSize = opts[0].BatchSize
	}
	return r.db.WithContext(ctx).CreateInBatches(entities, batchSize).Error
}

// Page 分页查询
func (r *gormRepository[T]) Page(ctx context.Context, page int, pageSize int) ([]T, int64, error) {
	var entities []T
	var total int64

	db := r.applyPreloads(r.db.WithContext(ctx))
	if err := db.Model(new(T)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 使用查询选项
	opts := options.NewQueryOptions(
		options.WithPage(page),
		options.WithPageSize(pageSize),
	)
	db = opts.ApplyQueryOptions(db)

	if err := db.Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

// BatchDelete 批量删除
func (r *gormRepository[T]) BatchDelete(ctx context.Context, ids []any, opts ...*options.DeleteOptions) error {
	return r.db.WithContext(ctx).Delete(new(T), ids).Error
}

// BatchUpdate 批量更新
func (r *gormRepository[T]) BatchUpdate(ctx context.Context, entities []T) error {
	return r.db.WithContext(ctx).Save(entities).Error
}

// Count 统计记录数
func (r *gormRepository[T]) Count(ctx context.Context, entity T) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(entity).Count(&count).Error
	return count, err
}

// Exists 检查记录是否存在
func (r *gormRepository[T]) Exists(ctx context.Context, query interface{}, args ...interface{}) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(new(T)).Where(query, args...).Count(&count).Error
	return count > 0, err
}

// Delete 删除实体
func (r *gormRepository[T]) Delete(ctx context.Context, entity T, opts ...*options.DeleteOptions) error {
	if len(opts) > 0 {
		if opts[0].Force {
			return r.db.WithContext(ctx).Unscoped().Delete(&entity).Error
		}
		if !opts[0].DeletedAt.IsZero() {
			r.db = r.db.Set("deleted_at", opts[0].DeletedAt)
		}
		if opts[0].DeletedBy != "" {
			r.db = r.db.Set("deleted_by", opts[0].DeletedBy)
		}
	}
	return r.db.WithContext(ctx).Delete(entity).Where(entity).Error
}

// DeleteById 根据ID删除
func (r *gormRepository[T]) DeleteById(ctx context.Context, id any, opts ...*options.DeleteOptions) error {
	// var entity T
	entity := NewModel[T]()
	// entity.Init()
	if err := entity.SetID(id); err != nil {
		return err
	}
	return r.Delete(ctx, entity, opts...)
}

// Update 更新实体
func (r *gormRepository[T]) Update(ctx context.Context, entity T, updateFields map[string]interface{}) error {
	// 如果没有提供更新字段，则使用整个实体进行更新
	if updateFields == nil {
		return r.db.WithContext(ctx).Updates(entity).Error
	}
	// 只更新指定字段
	return r.db.WithContext(ctx).Model(entity).Updates(updateFields).Error
}

// Preload 添加预加载
func (r *gormRepository[T]) Preload(query ...string) IRepository[T] {
	r.preloads = append(r.preloads, query...)
	return r
}

// WithTx 使用事务
func (r *gormRepository[T]) WithTx(tx *gorm.DB) IRepository[T] {
	return &gormRepository[T]{
		db:         tx,
		entityType: r.entityType,
		preloads:   r.preloads,
	}
}

// Transaction 事务操作
func (r *gormRepository[T]) Transaction(ctx context.Context, fc func(tx IRepository[T]) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fc(r.WithTx(tx))
	})
}
