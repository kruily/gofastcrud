package crud

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseUUIDEntity struct {
	BaseEntity[uuid.UUID]
}

// GetID 获取ID
func (e *BaseUUIDEntity) GetID() uuid.UUID {
	return e.ID
}

// SetID 设置ID
func (e *BaseUUIDEntity) SetID(id uuid.UUID) {
	e.ID = id
}

func (e *BaseUUIDEntity) BeforeCreate(tx *gorm.DB) error {
	if e == nil {
		e = &BaseUUIDEntity{}
	}
	e.ID = uuid.New()
	return nil
}
