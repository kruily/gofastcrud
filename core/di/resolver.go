package di

import (
	"fmt"
	"reflect"
)

// Resolve 解析依赖
func (c *Container) Resolve(iface interface{}) error {
	ifaceValue := reflect.ValueOf(iface)
	if ifaceValue.Kind() != reflect.Ptr {
		return fmt.Errorf("interface must be a pointer, got %v", ifaceValue.Kind())
	}

	ifaceType := ifaceValue.Type().Elem()
	instance, err := c.resolveType(ifaceType)
	if err != nil {
		return err
	}

	ifaceValue.Elem().Set(reflect.ValueOf(instance))
	return nil
}

// resolveType 解析类型
func (c *Container) resolveType(t reflect.Type) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 检查是否已有实例（单例模式）
	if instance, ok := c.instances[t]; ok {
		return instance, nil
	}

	// 获取绑定信息
	binding, ok := c.bindings[t]
	if !ok {
		return nil, fmt.Errorf("no binding found for type %v", t)
	}

	// 创建实例
	instance, err := c.createInstance(binding.implementation)
	if err != nil {
		return nil, err
	}

	// 如果是单例，保存实例
	if binding.singleton {
		c.instances[t] = instance
	}

	return instance, nil
}

// createInstance 创建实例
func (c *Container) createInstance(impl interface{}) (interface{}, error) {
	implType := reflect.TypeOf(impl)
	if implType.Kind() == reflect.Ptr {
		implType = implType.Elem()
	}

	// 创建新实例
	instance := reflect.New(implType).Interface()

	// 注入依赖
	if err := c.injectDependencies(instance); err != nil {
		return nil, err
	}

	return instance, nil
}
