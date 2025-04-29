package crud

import (
	"reflect"

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
