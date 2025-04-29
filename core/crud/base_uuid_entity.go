package crud

import (
	"time"

	"github.com/google/uuid"
	"github.com/kruily/gofastcrud/errors"
	"gorm.io/gorm"
)

type BaseUUIDEntity struct {
	ID        uuid.UUID `gorm:"type:string;primarykey;" json:"id" example:"1" description:"唯一标识符" filter:"eq,neq,in,nin"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at" example:"2024-03-20T10:00:00Z" description:"创建时间" filter:"gt,gte,lt,lte,eq,neq,in,nin"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at" example:"2024-03-20T10:00:00Z" description:"更新时间" filter:"gt,gte,lt,lte,eq,neq,in,nin"`
	DeleteAt  time.Time `gorm:"column:deleted_at;index" json:"-" example:"2024-03-20T10:00:00Z" description:"删除时间" filter:"gt,gte,lt,lte,eq,neq,in,nin"` // 软删除
}

// func ()

// GetID 获取ID
func (e *BaseUUIDEntity) GetID() any {
	if e == nil {
		e = &BaseUUIDEntity{}
	}
	return e.ID
}

// SetID 设置ID
func (e *BaseUUIDEntity) SetID(id any) error {
	if idUUID, ok := id.(uuid.UUID); ok {
		e.ID = idUUID
	} else {
		return errors.New(errors.ErrIDType, "invalid id type")
	}
	return nil
}

// GetCreatedAt 获取创建时间
func (e *BaseUUIDEntity) GetCreatedAt() time.Time {
	return e.CreatedAt
}

// GetUpdatedAt 获取更新时间
func (e *BaseUUIDEntity) GetUpdatedAt() time.Time {
	return e.UpdatedAt
}

// GetDeletedAt 获取删除时间
func (e *BaseUUIDEntity) GetDeletedAt() time.Time {
	return e.DeleteAt
}

// SetCreatedAt 设置创建时间
func (e *BaseUUIDEntity) SetCreatedAt(t time.Time) {
	e.CreatedAt = t
}

// SetUpdatedAt 设置更新时间
func (e *BaseUUIDEntity) SetUpdatedAt(t time.Time) {
	e.UpdatedAt = t
}

// SetDeletedAt 设置删除时间
func (e *BaseUUIDEntity) SetDeletedAt(t time.Time) {
	e.DeleteAt = t
}

func (e *BaseUUIDEntity) BeforeCreate(tx *gorm.DB) (err error) {
	e.ID = uuid.New()
	return nil
}

func (b *BaseUUIDEntity) DBType() string {
	return "mysql"
}
