package crud

import (
	"time"

	"github.com/google/uuid"
	"github.com/kruily/gofastcrud/pkg/errors"
)

type ID_TYPE interface {
	~uint64 | ~string | uuid.UUID
}

// ICrudEntity CRUD 实体接口
type ICrudEntity interface {
	// Table 获取表名
	Table() string
	// SetID 设置ID
	SetID(id any) error
	// GetID 获取ID
	GetID() any
}

// BaseEntity 基础实体
type BaseEntity struct {
	ID        uint64    `gorm:"primarykey" json:"id" example:"1" description:"唯一标识符"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at" example:"2024-03-20T10:00:00Z" description:"创建时间"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at" example:"2024-03-20T10:00:00Z" description:"更新时间"`
}

// GetID 获取ID
func (e *BaseEntity) GetID() any {
	return e.ID
}

// SetID 设置ID
func (e *BaseEntity) SetID(id any) error {
	if idInt, ok := id.(uint64); ok {
		e.ID = idInt
	} else {
		return errors.New(errors.ErrIDType, "invalid id type")
	}
	return nil
}
