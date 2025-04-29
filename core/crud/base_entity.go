package crud

import (
	"time"

	"github.com/kruily/gofastcrud/errors"
)

// ICrudEntity CRUD 实体接口
type ICrudEntity interface {
	// SetID 设置ID
	SetID(id any) error
	// GetID 获取ID
	GetID() any
	// GetCreatedAt 获取创建时间
	GetCreatedAt() time.Time
	// GetUpdatedAt 获取更新时间
	GetUpdatedAt() time.Time
	// GetDeletedAt 获取删除时间
	GetDeletedAt() time.Time
	// SetCreatedAt 设置创建时间
	SetCreatedAt(time.Time)
	// SetUpdatedAt 设置更新时间
	SetUpdatedAt(time.Time)
	// SetDeletedAt 设置删除时间
	SetDeletedAt(time.Time)
	// DBType 获取数据库类型
	DBType() string
	// Table 获取表名
	TableName() string
	// Init方法
	Init()
}

// BaseEntity 基础实体
type BaseEntity struct {
	ID        uint64    `gorm:"primarykey" json:"id" example:"1" description:"唯一标识符"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at" example:"2024-03-20T10:00:00Z" description:"创建时间"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at" example:"2024-03-20T10:00:00Z" description:"更新时间"`
	DeletedAt time.Time `gorm:"column:deleted_at;index" json:"-" example:"2024-03-20T10:00:00Z" description:"删除时间"` // 软删除
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

// GetCreatedAt 获取创建时间
func (e *BaseEntity) GetCreatedAt() time.Time {
	return e.CreatedAt
}

// GetUpdatedAt 获取更新时间
func (e *BaseEntity) GetUpdatedAt() time.Time {
	return e.UpdatedAt
}

// GetDeletedAt 获取删除时间
func (e *BaseEntity) GetDeletedAt() time.Time {
	return e.DeletedAt
}

// SetCreatedAt 设置创建时间
func (e *BaseEntity) SetCreatedAt(t time.Time) {
	e.CreatedAt = t
}

// SetUpdatedAt 设置更新时间
func (e *BaseEntity) SetUpdatedAt(t time.Time) {
	e.UpdatedAt = t
}

// SetDeletedAt 设置删除时间
func (e *BaseEntity) SetDeletedAt(t time.Time) {
	e.DeletedAt = t
}

func (b *BaseEntity) DBType() string {
	return DB_TYPE_GORM
}
