package crud

import (
	"context"
	"reflect"

	"github.com/kruily/gofastcrud/core/crud/options"
	"github.com/kruily/gofastcrud/core/database"
)

// IRepository 仓储接口
type IRepository[T ICrudEntity] interface {
	// 基础操作
	Create(ctx context.Context, entity T) error
	Update(ctx context.Context, entity T, updateFields map[string]interface{}) error
	Delete(ctx context.Context, entity T, opts ...*options.DeleteOptions) error
	DeleteById(ctx context.Context, id any, opts ...*options.DeleteOptions) error
	FindById(ctx context.Context, id any) (T, error)
	Find(ctx context.Context, entity T, opts *options.QueryOptions) ([]T, error)
	Count(ctx context.Context, entity T) (int64, error)

	// 批量操作
	BatchCreate(ctx context.Context, entities []T, opts ...*options.BatchOptions) error
	BatchUpdate(ctx context.Context, entities []T) error
	BatchDelete(ctx context.Context, ids []any, opts ...*options.DeleteOptions) error

	// 条件查询
	FindOne(ctx context.Context, query interface{}, args ...interface{}) (T, error)
	FindAll(ctx context.Context, query interface{}, args ...interface{}) ([]T, error)
	Exists(ctx context.Context, query interface{}, args ...interface{}) (bool, error)

	// 高级查询
	// Page(ctx context.Context, page int, pageSize int) ([]T, int64, error)
	// Where(query interface{}, args ...interface{}) IRepository[T]
	// Order(value interface{}) IRepository[T]
	// Select(query interface{}, args ...interface{}) IRepository[T]
	// Preload(query ...string) IRepository[T]
	// Joins(query string, args ...interface{}) IRepository[T]
	// Group(query string) IRepository[T]
	// Having(query interface{}, args ...interface{}) IRepository[T]

	// 事务操作
	Transaction(ctx context.Context, fc func(tx IRepository[T]) error) error
	// WithTx(tx *gorm.DB) IRepository[T]

	// 聚合操作
	// Sum(ctx context.Context, field string) (float64, error)
	// CountField(ctx context.Context, field string) (int64, error)
	// Max(ctx context.Context, field string) (float64, error)
	// Min(ctx context.Context, field string) (float64, error)
	// Avg(ctx context.Context, field string) (float64, error)

	// 锁
	// LockForUpdate() IRepository[T]
	// SharedLock() IRepository[T]

	// Session 创建新会话，避免污染原有查询
	// Session() IRepository[T]

	// 查询钩子
	// AddQueryHook(hook QueryHook) IRepository[T]
}

// Repository 仓储实现
type Repository[T ICrudEntity] struct {
	crudRepo   map[string]IRepository[T]
	entityType reflect.Type
}

// NewRepository 创建仓储实例
func NewRepository[T ICrudEntity](db *database.Database, entity T) *Repository[T] {
	entityType := reflect.TypeOf(entity)
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	repo := &Repository[T]{
		entityType: entityType,
		crudRepo:   make(map[string]IRepository[T]),
	}
	// 根据实体的DBType选择具体的仓储实现
	switch entity.DBType() {
	case DB_TYPE_GORM:
		// repo.gormRepo = newGormRepository[T](db.DB(), entity)
		repo.crudRepo[DB_TYPE_GORM] = newGormRepository(db.DB(), entity)
	case DB_TYPE_MONGODB:
		// repo.crudRepo[DB_TYPE_MONGODB] = newMongoRepository[T](db.MDB(), entity)
	}

	return repo
}

// FindOne 查询单个实体
func (r *Repository[T]) FindOne(ctx context.Context, query interface{}, args ...interface{}) (T, error) {
	entity := NewModel[T]()
	return r.crudRepo[entity.DBType()].FindOne(ctx, query, args...)
}

// Find 查询实体列表
func (r *Repository[T]) Find(ctx context.Context, entity T, opts *options.QueryOptions) ([]T, error) {
	return r.crudRepo[entity.DBType()].Find(ctx, entity, opts)
}

// FindAll 查询所有符合条件的记录
func (r *Repository[T]) FindAll(ctx context.Context, query interface{}, args ...interface{}) ([]T, error) {
	entity := NewModel[T]()
	return r.crudRepo[entity.DBType()].FindAll(ctx, query, args...)
}

// FindById 根据ID查询
func (r *Repository[T]) FindById(ctx context.Context, id any) (T, error) {
	entity := NewModel[T]()
	return r.crudRepo[entity.DBType()].FindById(ctx, id)
}

// Create 创建实体
func (r *Repository[T]) Create(ctx context.Context, entity T) error {
	return r.crudRepo[entity.DBType()].Create(ctx, entity)
}

// BatchCreate 批量创建
func (r *Repository[T]) BatchCreate(ctx context.Context, entities []T, opts ...*options.BatchOptions) error {
	return r.crudRepo[entities[0].DBType()].BatchCreate(ctx, entities, opts...)
}

// Transaction 事务操作
func (r *Repository[T]) Transaction(ctx context.Context, fc func(tx IRepository[T]) error) error {

	entity := NewModel[T]()
	return r.crudRepo[entity.DBType()].Transaction(ctx, fc)
}

// BatchDelete 批量删除
func (r *Repository[T]) BatchDelete(ctx context.Context, ids []any, opts ...*options.DeleteOptions) error {
	entity := NewModel[T]()
	return r.crudRepo[entity.DBType()].BatchDelete(ctx, ids, opts...)
}

// BatchUpdate 批量更新
func (r *Repository[T]) BatchUpdate(ctx context.Context, entities []T) error {
	return r.crudRepo[entities[0].DBType()].BatchUpdate(ctx, entities)
}

// Count 统计记录数
func (r *Repository[T]) Count(ctx context.Context, entity T) (int64, error) {
	entity = NewModel[T]()
	return r.crudRepo[entity.DBType()].Count(ctx, entity)
}

// Exists 检查记录是否存在
func (r *Repository[T]) Exists(ctx context.Context, query interface{}, args ...interface{}) (bool, error) {
	entity := NewModel[T]()
	return r.crudRepo[entity.DBType()].Exists(ctx, query, args...)
}

// Delete 删除实体
func (r *Repository[T]) Delete(ctx context.Context, entity T, opts ...*options.DeleteOptions) error {
	return r.crudRepo[entity.DBType()].Delete(ctx, entity, opts...)
}

// DeleteById 根据ID删除
func (r *Repository[T]) DeleteById(ctx context.Context, id any, opts ...*options.DeleteOptions) error {
	entity := NewModel[T]()
	if err := entity.SetID(id); err != nil {
		return err
	}
	return r.crudRepo[entity.DBType()].DeleteById(ctx, entity, opts...)
}

// Update 更新实体
func (r *Repository[T]) Update(ctx context.Context, entity T, updateFields map[string]interface{}) error {
	// 如果没有提供更新字段，则使用整个实体进行更新
	return r.crudRepo[entity.DBType()].Update(ctx, entity, updateFields)
}
