package crud

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/kruily/gofastcrud/core/crud/options"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// IRepository 仓储接口
type IRepository[T ICrudEntity[TID], TID ID_TYPE] interface {
	// 基础操作
	Create(ctx context.Context, entity *T) error
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, entity *T, opts ...*options.DeleteOptions) error
	DeleteById(ctx context.Context, id TID, opts ...*options.DeleteOptions) error
	FindById(ctx context.Context, id TID) (*T, error)
	Find(ctx context.Context, entity *T, opts *options.QueryOptions) ([]T, error)
	Count(ctx context.Context, entity *T) (int64, error)

	// 批量操作
	BatchCreate(ctx context.Context, entities []T, opts ...*options.BatchOptions) error
	BatchUpdate(ctx context.Context, entities []T) error
	BatchDelete(ctx context.Context, ids []TID, opts ...*options.DeleteOptions) error

	// 条件查询
	FindOne(ctx context.Context, query interface{}, args ...interface{}) (*T, error)
	FindAll(ctx context.Context, query interface{}, args ...interface{}) ([]T, error)
	Exists(ctx context.Context, query interface{}, args ...interface{}) (bool, error)

	// 高级查询
	Page(ctx context.Context, page int, pageSize int) ([]T, int64, error)
	Where(query interface{}, args ...interface{}) IRepository[T, TID]
	Order(value interface{}) IRepository[T, TID]
	Select(query interface{}, args ...interface{}) IRepository[T, TID]
	Preload(query ...string) IRepository[T, TID]
	Joins(query string, args ...interface{}) IRepository[T, TID]
	Group(query string) IRepository[T, TID]
	Having(query interface{}, args ...interface{}) IRepository[T, TID]

	// 事务操作
	Transaction(ctx context.Context, fc func(tx IRepository[T, TID]) error) error
	WithTx(tx *gorm.DB) IRepository[T, TID]

	// 聚合操作
	Sum(ctx context.Context, field string) (float64, error)
	CountField(ctx context.Context, field string) (int64, error)
	Max(ctx context.Context, field string) (float64, error)
	Min(ctx context.Context, field string) (float64, error)
	Avg(ctx context.Context, field string) (float64, error)

	// 锁
	LockForUpdate() IRepository[T, TID]
	SharedLock() IRepository[T, TID]

	// Session 创建新会话，避免污染原有查询
	Session() IRepository[T, TID]

	// 查询钩子
	AddQueryHook(hook QueryHook) IRepository[T, TID]
}

// Repository 仓储实现
type Repository[T ICrudEntity[TID], TID ID_TYPE] struct {
	db         *gorm.DB
	entityType reflect.Type
	preloads   []string // 预加载字段
}

// NewRepository 创建仓储实例
func NewRepository[T ICrudEntity[TID], TID ID_TYPE](db *gorm.DB, entity T) *Repository[T, TID] {
	entityType := reflect.TypeOf(entity)
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}
	return &Repository[T, TID]{
		db:         db,
		entityType: entityType,
		preloads:   make([]string, 0),
	}
}

// Session 创建新会话
func (r *Repository[T, TID]) Session() IRepository[T, TID] {
	return &Repository[T, TID]{
		db:         r.db.Session(&gorm.Session{}),
		entityType: r.entityType,
		preloads:   make([]string, 0),
	}
}

// AddQueryHook 添加查询钩子
func (r *Repository[T, TID]) AddQueryHook(hook QueryHook) IRepository[T, TID] {
	// 创建新的会话以避免污染原有查询
	db := r.db.Session(&gorm.Session{})
	// 注册回调
	db.Callback().Query().Before("gorm:query").Register("my_hook:before", hook.BeforeQuery)
	db.Callback().Query().After("gorm:query").Register("my_hook:after", hook.AfterQuery)
	r.db = db
	return r
}

// Preload 添加预加载
func (r *Repository[T, TID]) Preload(query ...string) IRepository[T, TID] {
	r.preloads = append(r.preloads, query...)
	return r
}

// applyPreloads 应用预加载
func (r *Repository[T, TID]) applyPreloads(db *gorm.DB) *gorm.DB {
	for _, preload := range r.preloads {
		db = db.Preload(preload)
	}
	return db
}

// WithTx 使用事务
func (r *Repository[T, TID]) WithTx(tx *gorm.DB) IRepository[T, TID] {
	return &Repository[T, TID]{
		db:         tx,
		entityType: r.entityType,
		preloads:   r.preloads,
	}
}

// FindOne 查询单个实体
func (r *Repository[T, TID]) FindOne(ctx context.Context, query interface{}, args ...interface{}) (*T, error) {
	var entity T
	db := r.applyPreloads(r.db.WithContext(ctx))
	err := db.Where(query, args...).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// Find 查询实体列表
func (r *Repository[T, TID]) Find(ctx context.Context, entity *T, opts *options.QueryOptions) ([]T, error) {
	var entities []T
	db := r.applyPreloads(r.db.WithContext(ctx).Model(entity))

	// 应用查询选项
	db = opts.ApplyQueryOptions(db)

	err := db.Find(&entities).Error
	return entities, err
}

// FindAll 查询所有符合条件的记录
func (r *Repository[T, TID]) FindAll(ctx context.Context, query interface{}, args ...interface{}) ([]T, error) {
	var entities []T
	db := r.applyPreloads(r.db.WithContext(ctx))
	err := db.Where(query, args...).Find(&entities).Error
	return entities, err
}

// FindById 根据ID查询
func (r *Repository[T, TID]) FindById(ctx context.Context, id TID) (*T, error) {
	var entity T
	db := r.applyPreloads(r.db.WithContext(ctx))
	err := db.First(&entity, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// 实现所有接口方法...
func (r *Repository[T, TID]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).
		Create(entity).Error
}

// BatchCreate 批量创建
func (r *Repository[T, TID]) BatchCreate(ctx context.Context, entities []T, opts ...*options.BatchOptions) error {
	batchSize := 100
	if len(opts) > 0 && opts[0].BatchSize > 0 {
		batchSize = opts[0].BatchSize
	}
	return r.db.WithContext(ctx).CreateInBatches(entities, batchSize).Error
}

// Page 分页查询
func (r *Repository[T, TID]) Page(ctx context.Context, page int, pageSize int) ([]T, int64, error) {
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

// Transaction 事务操作
func (r *Repository[T, TID]) Transaction(ctx context.Context, fc func(tx IRepository[T, TID]) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fc(r.WithTx(tx))
	})
}

// LockForUpdate 行锁
func (r *Repository[T, TID]) LockForUpdate() IRepository[T, TID] {
	return &Repository[T, TID]{
		db: r.db.Clauses(clause.Locking{Strength: "UPDATE"}),
	}
}

// Avg 计算平均值
func (r *Repository[T, TID]) Avg(ctx context.Context, field string) (float64, error) {
	var result float64
	err := r.db.WithContext(ctx).Model(new(T)).Select(fmt.Sprintf("AVG(%s)", field)).Scan(&result).Error
	return result, err
}

// Sum 计算总和
func (r *Repository[T, TID]) Sum(ctx context.Context, field string) (float64, error) {
	var result float64
	err := r.db.WithContext(ctx).Model(new(T)).Select(fmt.Sprintf("SUM(%s)", field)).Scan(&result).Error
	return result, err
}

// Max 计算最大值
func (r *Repository[T, TID]) Max(ctx context.Context, field string) (float64, error) {
	var result float64
	err := r.db.WithContext(ctx).Model(new(T)).Select(fmt.Sprintf("MAX(%s)", field)).Scan(&result).Error
	return result, err
}

// Min 计算最小值
func (r *Repository[T, TID]) Min(ctx context.Context, field string) (float64, error) {
	var result float64
	err := r.db.WithContext(ctx).Model(new(T)).Select(fmt.Sprintf("MIN(%s)", field)).Scan(&result).Error
	return result, err
}

// BatchDelete 批量删除
func (r *Repository[T, TID]) BatchDelete(ctx context.Context, ids []TID, opts ...*options.DeleteOptions) error {
	return r.db.WithContext(ctx).Delete(new(T), ids).Error
}

// BatchUpdate 批量更新
func (r *Repository[T, TID]) BatchUpdate(ctx context.Context, entities []T) error {
	return r.db.WithContext(ctx).Save(entities).Error
}

// SharedLock 共享锁
func (r *Repository[T, TID]) SharedLock() IRepository[T, TID] {
	return &Repository[T, TID]{
		db: r.db.Clauses(clause.Locking{Strength: "SHARE"}),
	}
}

// Count 统计记录数
func (r *Repository[T, TID]) Count(ctx context.Context, entity *T) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(entity).Count(&count).Error
	return count, err
}

// Exists 检查记录是否存在
func (r *Repository[T, TID]) Exists(ctx context.Context, query interface{}, args ...interface{}) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(new(T)).Where(query, args...).Count(&count).Error
	return count > 0, err
}

// CountField 统计指定字段的记录数
func (r *Repository[T, TID]) CountField(ctx context.Context, field string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(new(T)).Select(fmt.Sprintf("COUNT(%s)", field)).Scan(&count).Error
	return count, err
}

// Delete 删除实体
func (r *Repository[T, TID]) Delete(ctx context.Context, entity *T, opts ...*options.DeleteOptions) error {
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
func (r *Repository[T, TID]) DeleteById(ctx context.Context, id TID, opts ...*options.DeleteOptions) error {
	entity := new(T)
	return r.Delete(ctx, entity, opts...)
}

// 链式查询方法
func (r *Repository[T, TID]) Where(query interface{}, args ...interface{}) IRepository[T, TID] {
	r.db = r.db.Where(query, args...)
	return r
}

// Order 排序
func (r *Repository[T, TID]) Order(value interface{}) IRepository[T, TID] {
	r.db = r.db.Order(value)
	return r
}

// Select 选择字段
func (r *Repository[T, TID]) Select(query interface{}, args ...interface{}) IRepository[T, TID] {
	r.db = r.db.Select(query, args...)
	return r
}

// Joins 连接查询
func (r *Repository[T, TID]) Joins(query string, args ...interface{}) IRepository[T, TID] {
	r.db = r.db.Joins(query, args...)
	return r
}

// Group 分组
func (r *Repository[T, TID]) Group(query string) IRepository[T, TID] {
	r.db = r.db.Group(query)
	return r
}

// Having 过滤条件
func (r *Repository[T, TID]) Having(query interface{}, args ...interface{}) IRepository[T, TID] {
	r.db = r.db.Having(query, args...)
	return r
}

// Update 更新实体
func (r *Repository[T, TID]) Update(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Save(entity).Error
}
