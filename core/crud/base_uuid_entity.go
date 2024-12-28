package crud

import (
	"time"

	"github.com/google/uuid"
	"github.com/kruily/gofastcrud/pkg/errors"
	"gorm.io/gorm"
)

type BaseUUIDEntity struct {
	ID        uuid.UUID `gorm:"primarykey" json:"id" example:"1" description:"唯一标识符"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at" example:"2024-03-20T10:00:00Z" description:"创建时间"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at" example:"2024-03-20T10:00:00Z" description:"更新时间"`
}

// GetID 获取ID
func (e *BaseUUIDEntity) GetID() any {
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

func (e *BaseUUIDEntity) BeforeCreate(tx *gorm.DB) error {
	if e == nil {
		e = &BaseUUIDEntity{}
	}
	e.ID = uuid.New()
	return nil
}
