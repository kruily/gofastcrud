package fast_casbin

import (
	"errors"
	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

var (
	ErrNotInitialized   = errors.New("rbac enforcer not initialized")
	ErrNoPermission     = errors.New("permission denied")
	ErrInvalidModel     = errors.New("invalid model")
	ErrInvalidModelText = errors.New("model path or model text is required")
)

type CasbinMaker struct {
	enforcer *casbin.Enforcer
	mu       sync.RWMutex
}

// Config 配置
type Config struct {
	ModelText string   // Casbin模型配置文本
	ModelPath string   // Casbin模型文件路径
	DB        *gorm.DB // 数据库连接
	TableName string   // 规则表名称
}

// NewCasbinMaker 创建一个新的RBAC maker
func NewCasbinMaker(config Config) (*CasbinMaker, error) {
	if config.TableName == "" {
		config.TableName = "casbin_rules"
	}

	// 加载模型
	var m model.Model
	var err error
	if config.ModelPath != "" {
		m, err = model.NewModelFromFile(config.ModelPath)
	} else if config.ModelText != "" {
		m, err = model.NewModelFromString(config.ModelText)
	} else {
		m, err = model.NewModelFromFile("model.conf")
	}
	if err != nil {
		return nil, err
	}

	// 创建gorm适配器
	adapter, err := gormadapter.NewAdapterByDB(config.DB)
	if err != nil {
		return nil, err
	}

	// 创建enforcer
	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, err
	}

	// 加载策略
	if err := enforcer.LoadPolicy(); err != nil {
		return nil, err
	}

	return &CasbinMaker{
		enforcer: enforcer,
	}, nil
}

// AddPolicy 添加策略
func (rm *CasbinMaker) AddPolicy(sub, obj, act string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	_, err := rm.enforcer.AddPolicy(sub, obj, act)
	if err != nil {
		return err
	}
	return rm.enforcer.SavePolicy()
}

// AddPolicies 批量添加策略
func (rm *CasbinMaker) AddPolicies(rules [][]string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	_, err := rm.enforcer.AddPolicies(rules)
	if err != nil {
		return err
	}
	return rm.enforcer.SavePolicy()
}

// RemovePolicy 删除策略
func (rm *CasbinMaker) RemovePolicy(sub, obj, act string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	_, err := rm.enforcer.RemovePolicy(sub, obj, act)
	if err != nil {
		return err
	}
	return rm.enforcer.SavePolicy()
}

// AddRoleForUser 为用户添加角色
func (rm *CasbinMaker) AddRoleForUser(user, role string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	_, err := rm.enforcer.AddGroupingPolicy(user, role)
	if err != nil {
		return err
	}
	return rm.enforcer.SavePolicy()
}

// RemoveRoleForUser 删除用户的角色
func (rm *CasbinMaker) RemoveRoleForUser(user, role string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	_, err := rm.enforcer.RemoveGroupingPolicy(user, role)
	if err != nil {
		return err
	}
	return rm.enforcer.SavePolicy()
}

// Enforce 检查权限
func (rm *CasbinMaker) Enforce(sub, obj, act string) (bool, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if rm.enforcer == nil {
		return false, ErrNotInitialized
	}

	return rm.enforcer.Enforce(sub, obj, act)
}

// GetRolesForUser 获取用户的所有角色
func (rm *CasbinMaker) GetRolesForUser(user string) ([]string, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return rm.enforcer.GetRolesForUser(user)
}

// GetAllRoles 获取所有角色
func (rm *CasbinMaker) GetAllRoles() ([]string, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return rm.enforcer.GetAllRoles()
}

// GetUsersForRole 获取具有指定角色的所有用户
func (rm *CasbinMaker) GetUsersForRole(role string) ([]string, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return rm.enforcer.GetUsersForRole(role)
}

// HasRoleForUser 检查用户是否具有指定角色
func (rm *CasbinMaker) HasRoleForUser(user, role string) (bool, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return rm.enforcer.HasRoleForUser(user, role)
}

// InitPolicy 初始化默认策略
func (rm *CasbinMaker) InitPolicy() error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// 创建超级管理员角色
	if _, err := rm.enforcer.AddPolicy("admin", "*", "*"); err != nil {
		return err
	}

	// 创建基本角色和权限
	policies := [][]string{
		{"user", "articles", "read"},
		{"editor", "articles", "write"},
		{"editor", "articles", "read"},
		{"manager", "users", "read"},
		{"manager", "users", "write"},
	}

	_, err := rm.enforcer.AddPolicies(policies)
	return err
}

// LoadPoliciesFromDB 从数据库重新加载策略
func (rm *CasbinMaker) LoadPoliciesFromDB() error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	return rm.enforcer.LoadPolicy()
}

// GetUserPermissions 获取用户所有权限
func (rm *CasbinMaker) GetUserPermissions(username string) ([][]string, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// 获取用户的所有角色
	roles, err := rm.enforcer.GetRolesForUser(username)
	if err != nil {
		return nil, err
	}

	// 获取所有权限
	permissions := make([][]string, 0)
	for _, role := range roles {
		ps, _ := rm.enforcer.GetPermissionsForUser(role)
		permissions = append(permissions, ps...)
	}

	return permissions, nil
}

// HasPermissionForUser 检查用户是否有特定权限
func (rm *CasbinMaker) HasPermissionForUser(username, obj, act string) bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// 检查直接权限
	if ok, _ := rm.enforcer.Enforce(username, obj, act); ok {
		return true
	}

	// 获取用户的所有角色
	roles, _ := rm.enforcer.GetRolesForUser(username)
	for _, role := range roles {
		// 检查角色权限
		if ok, _ := rm.enforcer.Enforce(role, obj, act); ok {
			return true
		}
	}

	return false
}

// AddUserRole 为用户添加角色
func (rm *CasbinMaker) AddUserRole(username, role string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	_, err := rm.enforcer.AddGroupingPolicy(username, role)
	return err
}

// RemoveUserRole 删除用户的角色
func (rm *CasbinMaker) RemoveUserRole(username, role string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	_, err := rm.enforcer.RemoveGroupingPolicy(username, role)
	return err
}
