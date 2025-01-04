package fast_casbin

// CasbinRule 存储在数据库中的Casbin规则
type CasbinRule struct {
	ID    uint   `gorm:"primarykey"`
	Ptype string `gorm:"size:100;not null"` // 策略类型：p（权限）或 g（角色）
	V0    string `gorm:"size:100;not null"` // 对于p: 角色；对于g: 用户
	V1    string `gorm:"size:100;not null"` // 对于p: 资源；对于g: 角色
	V2    string `gorm:"size:100"`          // 对于p: 操作；对于g: 域（可选）
	V3    string `gorm:"size:100"`          // 额外字段
	V4    string `gorm:"size:100"`          // 额外字段
	V5    string `gorm:"size:100"`          // 额外字段
}

// TableName 指定表名
func (CasbinRule) TableName() string {
	return "casbin_rules"
}
