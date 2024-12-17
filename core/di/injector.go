package di

import (
	"fmt"
	"reflect"
)

// injectDependencies 注入依赖
func (c *Container) injectDependencies(instance interface{}) error {
	val := reflect.ValueOf(instance)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("instance must be a pointer")
	}

	val = val.Elem()
	typ := val.Type()

	// 遍历所有字段
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// 检查是否有 `inject` 标签
		if _, ok := field.Tag.Lookup("inject"); !ok {
			continue
		}

		// 字段必须是指针类型
		if !fieldValue.CanSet() || field.Type.Kind() != reflect.Ptr {
			continue
		}

		// 解析依赖
		dependency, err := c.resolveType(field.Type.Elem())
		if err != nil {
			return fmt.Errorf("failed to inject field %s: %v", field.Name, err)
		}

		// 设置依赖
		fieldValue.Set(reflect.ValueOf(dependency))
	}

	return nil
}

// MustResolve 解析依赖（如果出错则panic）
func (c *Container) MustResolve(iface interface{}) {
	if err := c.Resolve(iface); err != nil {
		panic(err)
	}
}
