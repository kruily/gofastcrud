package crud

import (
	"context"
	"errors"
	"time"

	"github.com/qiniu/qmgo/field"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BaseMongoEntity struct {
	Id        primitive.ObjectID `bson:"_id,omitempty" json:"id" `
	CreatedAt time.Time          `bson:"created_at" json:"created_at" example:"2024-03-20T10:00:00Z" description:"创建时间"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at" example:"2024-03-20T10:00:00Z" description:"更新时间"`
	DeletedAt time.Time          `bson:"deleted_at" json:"-" example:"2024-03-20T10:00:00Z" description:"删除时间"`
}

func (b *BaseMongoEntity) GetID() any {
	return b.Id
}

func (b *BaseMongoEntity) SetID(id any) error {
	if _, ok := id.(primitive.ObjectID); ok {
		b.Id = id.(primitive.ObjectID)
		return nil
	}
	return errors.New("id is invalid")
}

func (b *BaseMongoEntity) GetCreatedAt() time.Time {
	return b.CreatedAt
}

func (b *BaseMongoEntity) SetCreatedAt(createdAt time.Time) {
	b.CreatedAt = createdAt
}

func (b *BaseMongoEntity) GetUpdatedAt() time.Time {
	return b.UpdatedAt
}

func (b *BaseMongoEntity) SetUpdatedAt(updatedAt time.Time) {
	b.UpdatedAt = updatedAt
}

func (b *BaseMongoEntity) GetDeletedAt() time.Time {
	return b.DeletedAt
}

func (b *BaseMongoEntity) SetDeletedAt(deletedAt time.Time) {
	b.DeletedAt = deletedAt
}

func (b *BaseMongoEntity) DBType() string {
	return DB_TYPE_MONGODB
}

// 指定自定义field的field名
func (b *BaseMongoEntity) CustomFields() field.CustomFieldsBuilder {
	return field.NewCustom().SetCreateAt("CreateAt").
		SetUpdateAt("UpdateAt").
		SetId("Id")
}

// 在结构体中实现钩子接口（qmgo示例）
func (b *BaseMongoEntity) BeforeInsert(ctx context.Context) error {
	if b.CreatedAt.IsZero() {
		b.CreatedAt = time.Now()
	}
	b.UpdatedAt = time.Now()
	return nil
}

func (b *BaseMongoEntity) BeforeUpdate(ctx context.Context) error {
	b.UpdatedAt = time.Now()
	return nil
}
