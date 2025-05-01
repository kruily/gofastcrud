package crud

import (
	"context"
	"errors"
	"reflect"

	"github.com/kruily/gofastcrud/core/crud/options"
	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// mongoRepository mongodb仓储实现
type mongoRepository[T ICrudEntity] struct {
	collection *qmgo.Collection
	entityType reflect.Type
}

// newMongoRepository 创建mongodb仓储实例
func newMongoRepository[T ICrudEntity](db *qmgo.Database, entity T) *mongoRepository[T] {
	entityType := reflect.TypeOf(entity)
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}
	return &mongoRepository[T]{
		collection: db.Collection(entity.TableName()), // 不确定是不是使用entity的TableName方法
		entityType: entityType,
	}
}

func (r *mongoRepository[T]) Create(ctx context.Context, entity T) error {
	res, err := r.collection.InsertOne(ctx, entity)
	if err == nil {
		entity.SetID(res.InsertedID)
	}
	return err
}

func (r *mongoRepository[T]) Update(ctx context.Context, entity T, updateFields map[string]interface{}) error {
	if len(updateFields) == 0 {
		return r.collection.UpdateOne(ctx, entity, entity)
	}
	err := r.collection.UpdateOne(ctx, entity, updateFields) // TODO 不确定是否如此写
	return err
}
func (r *mongoRepository[T]) Delete(ctx context.Context, entity T, opts ...*options.DeleteOptions) error {
	err := r.collection.Remove(ctx, entity)
	return err
}
func (r *mongoRepository[T]) DeleteById(ctx context.Context, id any, opts ...*options.DeleteOptions) error {
	err := r.collection.RemoveId(ctx, id)
	return err
}
func (r *mongoRepository[T]) FindById(ctx context.Context, id any) (T, error) {
	entity := NewModel[T]()
	objId, err := Id2ObjectId(id)
	if err != nil {
		return entity, err
	}
	if err := entity.SetID(objId); err != nil {
		return entity, err
	}
	err = r.collection.Find(ctx, bson.D{{Key: "_id", Value: objId}}).One(entity)
	return entity, err
}
func (r *mongoRepository[T]) Find(ctx context.Context, entity T, opts *options.QueryOptions) ([]T, error) {
	var entities []T
	err := r.collection.Find(ctx, entity).All(entities)
	return entities, err
}

func (r *mongoRepository[T]) Count(ctx context.Context, entity T) (int64, error) {
	return r.collection.Find(ctx, entity).Count()
}

func (r *mongoRepository[T]) BatchCreate(ctx context.Context, entities []T, opts ...*options.BatchOptions) error {
	_, err := r.collection.InsertMany(ctx, entities)
	return err
}

func (r *mongoRepository[T]) BatchUpdate(ctx context.Context, entities []T) error {
	_, err := r.collection.UpdateAll(ctx, entities, entities)
	return err
}
func (r *mongoRepository[T]) BatchDelete(ctx context.Context, ids []any, opts ...*options.DeleteOptions) error {
	err := r.collection.Remove(ctx, bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: ids}}}})
	return err
}
func (r *mongoRepository[T]) FindOne(ctx context.Context, query interface{}, args ...interface{}) (T, error) {
	entity := NewModel[T]()
	err := r.collection.Find(ctx, query).One(entity)
	return entity, err
}
func (r *mongoRepository[T]) FindAll(ctx context.Context, query interface{}, args ...interface{}) ([]T, error) {
	var entities []T
	err := r.collection.Find(ctx, query).All(entities)
	if err != nil {
		return nil, err
	}
	return entities, err
}
func (r *mongoRepository[T]) Exists(ctx context.Context, query interface{}, args ...interface{}) (bool, error) {
	entity := NewModel[T]()
	err := r.collection.Find(ctx, query).One(entity)
	return entity.GetID() != nil, err
}
func (r *mongoRepository[T]) Transaction(ctx context.Context, fc func(tx IRepository[T]) error) error {
	// _, err := r.collection.Aggregate(ctx,fc(),)// TODO 待完善
	return nil
}

func Id2ObjectId(id any) (primitive.ObjectID, error) {
	if id == nil {
		return primitive.NilObjectID, nil
	}
	if idStr, ok := id.(string); ok {
		return primitive.ObjectIDFromHex(idStr)
	}
	return primitive.NilObjectID, errors.New("id must be string")
}
