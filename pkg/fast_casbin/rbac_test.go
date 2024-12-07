package fast_casbin

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const testModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	return db
}

func TestCasbinMaker(t *testing.T) {
	db := setupTestDB(t)

	config := Config{
		ModelText: testModel,
		DB:        db,
		TableName: "casbin_rules",
	}

	maker, err := NewCasbinMaker(config)
	require.NoError(t, err)

	// 测试添加角色和权限
	err = maker.AddPolicy("admin", "article", "write")
	require.NoError(t, err)

	err = maker.AddRoleForUser("alice", "admin")
	require.NoError(t, err)

	// 测试权限检查
	allowed, err := maker.Enforce("alice", "article", "write")
	require.NoError(t, err)
	require.True(t, allowed)

	// 测试无权限的情况
	allowed, err = maker.Enforce("bob", "article", "write")
	require.NoError(t, err)
	require.False(t, allowed)

	// 测试获取用户角色
	roles, err := maker.GetRolesForUser("alice")
	require.NoError(t, err)
	require.Contains(t, roles, "admin")

	// 测试删除角色
	err = maker.RemoveRoleForUser("alice", "admin")
	require.NoError(t, err)

	allowed, err = maker.Enforce("alice", "article", "write")
	require.NoError(t, err)
	require.False(t, allowed)

	// 测试批量添加策略
	rules := [][]string{
		{"editor", "article", "read"},
		{"editor", "article", "write"},
	}
	err = maker.AddPolicies(rules)
	require.NoError(t, err)

	// 测试获取所有角色
	allRoles, err := maker.GetAllRoles()
	require.NoError(t, err)
	require.Contains(t, allRoles, "admin")
	require.Contains(t, allRoles, "editor")

	// 测试获取角色的用户
	err = maker.AddRoleForUser("bob", "editor")
	require.NoError(t, err)

	users, err := maker.GetUsersForRole("editor")
	require.NoError(t, err)
	require.Contains(t, users, "bob")

	// 测试检查用户角色
	hasRole, err := maker.HasRoleForUser("bob", "editor")
	require.NoError(t, err)
	require.True(t, hasRole)
}
