package crud

import (
	"context"
	"reflect"

	"github.com/kruily/gofastcrud/core/crud/options"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// mongoRepository mongodb仓储实现
type mongoRepository[T ICrudEntity] struct {
	db         *mongo.Database
	collection *mongo.Collection
	entityType reflect.Type
}

// newMongoRepository 创建mongodb仓储实例
func newMongoRepository[T ICrudEntity](db *mongo.Database, entity T) *mongoRepository[T] {
	entityType := reflect.TypeOf(entity)
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}
	return &mongoRepository[T]{
		db:         db,
		collection: db.Collection(entity.TableName()), // 不确定是不是使用entity的TableName方法
		entityType: entityType,
	}
}

func (r *mongoRepository[T]) Create(ctx context.Context, entity T) error {
	_, err := r.collection.InsertOne(ctx, entity)
	return err
}
func (r *mongoRepository[T]) Update(ctx context.Context, entity T, updateFields map[string]interface{}) error {
	_, err := r.collection.UpdateOne(ctx, entity, updateFields) // TODO 不确定是否如此写
	return err
}
func (r *mongoRepository[T]) Delete(ctx context.Context, entity T, opts ...*options.DeleteOptions) error {
	_, err := r.collection.DeleteOne(ctx, entity)
	return err
}
func (r *mongoRepository[T]) DeleteById(ctx context.Context, id any, opts ...*options.DeleteOptions) error {
	entity := NewModel[T]()
	if err := entity.SetID(id); err != nil {
		return err
	}
	_, err := r.collection.DeleteOne(ctx, entity)
	// _,err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	return err
}
func (r *mongoRepository[T]) FindById(ctx context.Context, id any) (T, error) {
	entity := NewModel[T]()
	if err := entity.SetID(id); err != nil {
		return entity, err
	}
	err := r.collection.FindOne(ctx, entity).Decode(entity)
	// err := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(entity)
	return entity, err
}
func (r *mongoRepository[T]) Find(ctx context.Context, entity T, opts *options.QueryOptions) ([]T, error) {
	cursor, err := r.collection.Find(ctx, entity)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var entities []T
	for cursor.Next(ctx) {
		var entity T
		if err := cursor.Decode(&entity); err != nil {
			return nil, err
		}
		entities = append(entities, entity)
	}
	return entities, nil
}
func (r *mongoRepository[T]) Count(ctx context.Context, entity T) (int64, error) {
	return r.collection.CountDocuments(ctx, entity)
}
func (r *mongoRepository[T]) BatchCreate(ctx context.Context, entities []T, opts ...*options.BatchOptions) error {
	es := make([]interface{}, len(entities))
	for i, entity := range entities {
		es[i] = entity
	}
	_, err := r.collection.InsertMany(ctx, es)
	return err
}
func (r *mongoRepository[T]) BatchUpdate(ctx context.Context, entities []T) error {
	// TODO
	return nil
}
func (r *mongoRepository[T]) BatchDelete(ctx context.Context, ids []any, opts ...*options.DeleteOptions) error {
	_, err := r.collection.DeleteMany(ctx, bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: ids}}}})
	return err
}
func (r *mongoRepository[T]) FindOne(ctx context.Context, query interface{}, args ...interface{}) (T, error) {
	entity := NewModel[T]()
	err := r.collection.FindOne(ctx, query).Decode(entity)
	return entity, err
}
func (r *mongoRepository[T]) FindAll(ctx context.Context, query interface{}, args ...interface{}) ([]T, error) {
	cursor, err := r.collection.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var entities []T
	for cursor.Next(ctx) {
		var entity T
		if err := cursor.Decode(&entity); err != nil {
			return nil, err
		}
		entities = append(entities, entity)
	}
	return entities, nil
}
func (r *mongoRepository[T]) Exists(ctx context.Context, query interface{}, args ...interface{}) (bool, error) {
	res := r.collection.FindOne(ctx, query)
	return res.Err() == nil, res.Err()
}
func (r *mongoRepository[T]) Transaction(ctx context.Context, fc func(tx IRepository[T]) error) error {
	_, err := r.collection.Aggregate(ctx, bson.A{}) // TODO 待完善
	return err
}
