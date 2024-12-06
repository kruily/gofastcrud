package crud

import "time"

// ICrudEntity CRUD 实体接口
type ICrudEntity interface {
	// Table 获取表名
	Table() string
	// GetID 获取ID
	GetID() uint
	// SetID 设置ID
	SetID(id uint)
}

// BaseEntity 基础实体
type BaseEntity struct {
	ID        uint      `json:"id" example:"1" description:"唯一标识符"`
	CreatedAt time.Time `json:"created_at" example:"2024-03-20T10:00:00Z" description:"创建时间"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-03-20T10:00:00Z" description:"更新时间"`
}

// GetID 获取ID
func (e *BaseEntity) GetID() uint {
	return e.ID
}

// SetID 设置ID
func (e *BaseEntity) SetID(id uint) {
	e.ID = id
}
