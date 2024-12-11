package module

type ICasbin interface {
	IModule
	// GetEnforcer 获取 Casbin enforcer
	GetEnforcer() interface{}

	// AddPolicy 添加策略
	AddPolicy(sec string, ptype string, rules []string) (bool, error)

	// RemovePolicy 删除策略
	RemovePolicy(sec string, ptype string, rules []string) (bool, error)

	// HasPolicy 检查策略是否存在
	HasPolicy(sec string, ptype string, rules []string) bool

	// GetPolicy 获取所有策略
	GetPolicy(sec string, ptype string) [][]string

	// GetFilteredPolicy 获取过滤后的策略
	GetFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) [][]string

	// AddGroupingPolicy 添加组策略
	AddGroupingPolicy(rules []string) (bool, error)

	// RemoveGroupingPolicy 删除组策略
	RemoveGroupingPolicy(rules []string) (bool, error)

	// HasGroupingPolicy 检查组策略是否存在
	HasGroupingPolicy(rules []string) bool

	// GetGroupingPolicy 获取所有组策略
	GetGroupingPolicy() [][]string

	// GetFilteredGroupingPolicy 获取过滤后的组策略
	GetFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) [][]string

	// Enforce 执行权限检查
	Enforce(rvals ...interface{}) (bool, error)
}
