package crud

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// IRepository 仓储接口
type IRepository[T ICrudEntity] interface {
	// 基础操作
	Create(ctx context.Context, entity *T) error
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, entity *T, opts ...DeleteOptions) error
	DeleteById(ctx context.Context, id uint, opts ...DeleteOptions) error
	FindById(ctx context.Context, id uint) (*T, error)
	Find(ctx context.Context, entity *T, opts QueryOptions) ([]T, error)
	Count(ctx context.Context, entity *T) (int64, error)

	// 批量操作
	BatchCreate(ctx context.Context, entities []T, opts ...BatchOptions) error
	BatchUpdate(ctx context.Context, entities []T) error
	BatchDelete(ctx context.Context, ids []uint, opts ...DeleteOptions) error

	// 条件查询
	FindOne(ctx context.Context, query interface{}, args ...interface{}) (*T, error)
	FindAll(ctx context.Context, query interface{}, args ...interface{}) ([]T, error)
	Exists(ctx context.Context, query interface{}, args ...interface{}) (bool, error)

	// 高级查询
	Page(ctx context.Context, page int, pageSize int) ([]T, int64, error)
	Where(query interface{}, args ...interface{}) IRepository[T]
	Order(value interface{}) IRepository[T]
	Select(query interface{}, args ...interface{}) IRepository[T]
	Preload(query string, args ...interface{}) IRepository[T]
	Joins(query string, args ...interface{}) IRepository[T]
	Group(query string) IRepository[T]
	Having(query interface{}, args ...interface{}) IRepository[T]

	// 事务操作
	Transaction(ctx context.Context, fc func(tx IRepository[T]) error) error
	WithTx(tx *gorm.DB) IRepository[T]

	// 聚合操作
	Sum(ctx context.Context, field string) (float64, error)
	CountField(ctx context.Context, field string) (int64, error)
	Max(ctx context.Context, field string) (float64, error)
	Min(ctx context.Context, field string) (float64, error)
	Avg(ctx context.Context, field string) (float64, error)

	// 锁
	LockForUpdate() IRepository[T]
	SharedLock() IRepository[T]

	// Session 创建新会话，避免污染原有查询
	Session() IRepository[T]

	// QueryHook 查询钩子
	AddQueryHook(hook QueryHook) IRepository[T]
}

// Repository 仓储实现
type Repository[T ICrudEntity] struct {
	db *gorm.DB
}

// 实现所有接口方法...
func (r *Repository[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// BatchCreate 批量创建
func (r *Repository[T]) BatchCreate(ctx context.Context, entities []T, opts ...BatchOptions) error {
	batchSize := 100
	if len(opts) > 0 && opts[0].BatchSize > 0 {
		batchSize = opts[0].BatchSize
	}
	return r.db.WithContext(ctx).CreateInBatches(entities, batchSize).Error
}

// FindOne 查询单个实体
func (r *Repository[T]) FindOne(ctx context.Context, query interface{}, args ...interface{}) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).Where(query, args...).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// Page 分页查询
func (r *Repository[T]) Page(ctx context.Context, page int, pageSize int) ([]T, int64, error) {
	var entities []T
	var total int64

	db := r.db.WithContext(ctx)
	if err := db.Model(new(T)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

// Transaction 事务操作
func (r *Repository[T]) Transaction(ctx context.Context, fc func(tx IRepository[T]) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fc(r.WithTx(tx))
	})
}

// LockForUpdate 行锁
func (r *Repository[T]) LockForUpdate() IRepository[T] {
	return &Repository[T]{
		db: r.db.Clauses(clause.Locking{Strength: "UPDATE"}),
	}
}

// WithTx 使用事务
func (r *Repository[T]) WithTx(tx *gorm.DB) IRepository[T] {
	return &Repository[T]{db: tx}
}

// Avg 计算平均值
func (r *Repository[T]) Avg(ctx context.Context, field string) (float64, error) {
	var result float64
	err := r.db.WithContext(ctx).Model(new(T)).Select(fmt.Sprintf("AVG(%s)", field)).Scan(&result).Error
	return result, err
}

// Sum 计算总和
func (r *Repository[T]) Sum(ctx context.Context, field string) (float64, error) {
	var result float64
	err := r.db.WithContext(ctx).Model(new(T)).Select(fmt.Sprintf("SUM(%s)", field)).Scan(&result).Error
	return result, err
}

// Max 计算最大值
func (r *Repository[T]) Max(ctx context.Context, field string) (float64, error) {
	var result float64
	err := r.db.WithContext(ctx).Model(new(T)).Select(fmt.Sprintf("MAX(%s)", field)).Scan(&result).Error
	return result, err
}

// Min 计算最小值
func (r *Repository[T]) Min(ctx context.Context, field string) (float64, error) {
	var result float64
	err := r.db.WithContext(ctx).Model(new(T)).Select(fmt.Sprintf("MIN(%s)", field)).Scan(&result).Error
	return result, err
}

// BatchDelete 批量删除
func (r *Repository[T]) BatchDelete(ctx context.Context, ids []uint, opts ...DeleteOptions) error {
	return r.db.WithContext(ctx).Delete(new(T), ids).Error
}

// BatchUpdate 批量更新
func (r *Repository[T]) BatchUpdate(ctx context.Context, entities []T) error {
	return r.db.WithContext(ctx).Save(entities).Error
}

// SharedLock 共享锁
func (r *Repository[T]) SharedLock() IRepository[T] {
	return &Repository[T]{
		db: r.db.Clauses(clause.Locking{Strength: "SHARE"}),
	}
}

// Count 统计记录数
func (r *Repository[T]) Count(ctx context.Context, entity *T) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(entity).Count(&count).Error
	return count, err
}

// Exists 检查记录是否存在
func (r *Repository[T]) Exists(ctx context.Context, query interface{}, args ...interface{}) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(new(T)).Where(query, args...).Count(&count).Error
	return count > 0, err
}

// CountField 统计指定字段的记录数
func (r *Repository[T]) CountField(ctx context.Context, field string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(new(T)).Select(fmt.Sprintf("COUNT(%s)", field)).Scan(&count).Error
	return count, err
}

// Delete 删除实体
func (r *Repository[T]) Delete(ctx context.Context, entity *T, opts ...DeleteOptions) error {
	if len(opts) > 0 {
		if opts[0].Force {
			return r.db.WithContext(ctx).Unscoped().Delete(entity).Error
		}
		if !opts[0].DeletedAt.IsZero() {
			r.db = r.db.Set("deleted_at", opts[0].DeletedAt)
		}
		if opts[0].DeletedBy != "" {
			r.db = r.db.Set("deleted_by", opts[0].DeletedBy)
		}
	}
	return r.db.WithContext(ctx).Delete(entity).Error
}

// DeleteById 根据ID删除
func (r *Repository[T]) DeleteById(ctx context.Context, id uint, opts ...DeleteOptions) error {
	entity := new(T)
	return r.Delete(ctx, entity, opts...)
}

// Find 查询实体列表
func (r *Repository[T]) Find(ctx context.Context, entity *T, opts QueryOptions) ([]T, error) {
	var entities []T
	db := r.db.WithContext(ctx).Model(entity)

	// 应用查询选项
	if opts.Search != "" && len(opts.SearchFields) > 0 {
		for _, field := range opts.SearchFields {
			db = db.Or(field+" LIKE ?", "%"+opts.Search+"%")
		}
	}

	if len(opts.OrderBy) > 0 {
		for _, order := range opts.OrderBy {
			db = db.Order(order)
		}
	}

	if len(opts.Preload) > 0 {
		for _, preload := range opts.Preload {
			db = db.Preload(preload)
		}
	}

	offset := (opts.Page - 1) * opts.PageSize
	err := db.Offset(offset).Limit(opts.PageSize).Find(&entities).Error
	return entities, err
}

// FindAll 查询所有符合条件的记录
func (r *Repository[T]) FindAll(ctx context.Context, query interface{}, args ...interface{}) ([]T, error) {
	var entities []T
	err := r.db.WithContext(ctx).Where(query, args...).Find(&entities).Error
	return entities, err
}

// FindById 根据ID查询
func (r *Repository[T]) FindById(ctx context.Context, id uint) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).First(&entity, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// 链式查询方法
func (r *Repository[T]) Where(query interface{}, args ...interface{}) IRepository[T] {
	r.db = r.db.Where(query, args...)
	return r
}

// Order 排序
func (r *Repository[T]) Order(value interface{}) IRepository[T] {
	r.db = r.db.Order(value)
	return r
}

// Select 选择字段
func (r *Repository[T]) Select(query interface{}, args ...interface{}) IRepository[T] {
	r.db = r.db.Select(query, args...)
	return r
}

// Preload 预加载
func (r *Repository[T]) Preload(query string, args ...interface{}) IRepository[T] {
	r.db = r.db.Preload(query, args...)
	return r
}

// Joins 连接查询
func (r *Repository[T]) Joins(query string, args ...interface{}) IRepository[T] {
	r.db = r.db.Joins(query, args...)
	return r
}

// Group 分组
func (r *Repository[T]) Group(query string) IRepository[T] {
	r.db = r.db.Group(query)
	return r
}

// Having 过滤条件
func (r *Repository[T]) Having(query interface{}, args ...interface{}) IRepository[T] {
	r.db = r.db.Having(query, args...)
	return r
}

// Update 更新实体
func (r *Repository[T]) Update(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// NewRepository 创建仓储实例
func NewRepository[T ICrudEntity](db *gorm.DB, entity T) IRepository[T] {
	return &Repository[T]{
		db: db,
	}
}

// Session 创建新会话，避免污染原有查询
func (r *Repository[T]) Session() IRepository[T] {
	return &Repository[T]{
		db: r.db.Session(&gorm.Session{}),
	}
}

func (r *Repository[T]) AddQueryHook(hook QueryHook) IRepository[T] {
	// 创建新的会话以避免污染原有查询
	db := r.db.Session(&gorm.Session{})
	// 注册回调
	db.Callback().Query().Before("gorm:query").Register("my_hook:before", hook.BeforeQuery)
	db.Callback().Query().After("gorm:query").Register("my_hook:after", hook.AfterQuery)
	r.db = db
	return r
}
