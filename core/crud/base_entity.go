package crud

import (
	"time"

	"github.com/google/uuid"
)

type ID_TYPE interface {
	~uint64 | ~string | uuid.UUID
}

// ICrudEntity CRUD 实体接口
type ICrudEntity[T ID_TYPE] interface {
	// Table 获取表名
	Table() string
	// SetID 设置ID
	SetID(id T)
	// GetID 获取ID
	GetID() T
}

// BaseEntity 基础实体
type BaseEntity[T ID_TYPE] struct {
	ID        T         `gorm:"primarykey" json:"id" example:"1" description:"唯一标识符"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at" example:"2024-03-20T10:00:00Z" description:"创建时间"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at" example:"2024-03-20T10:00:00Z" description:"更新时间"`
}

// GetID 获取ID
func (e *BaseEntity[T]) GetID() T {
	return e.ID
}

// SetID 设置ID
func (e *BaseEntity[T]) SetID(id T) {
	e.ID = id
}
