package di

import (
	"fmt"
	"reflect"
	"sync"
)

// Container 依赖注入容器
type Container struct {
	mu        sync.RWMutex
	instances map[reflect.Type]interface{}
	bindings  map[reflect.Type]binding
}

// binding 绑定信息
type binding struct {
	implementation interface{}
	singleton      bool
}

// New 创建容器实例
func New() *Container {
	return &Container{
		instances: make(map[reflect.Type]interface{}),
		bindings:  make(map[reflect.Type]binding),
	}
}

// Bind 绑定接口到实现
func (c *Container) Bind(iface interface{}, impl interface{}) error {
	return c.BindWithOptions(iface, impl, false)
}

// BindSingleton 绑定单例
func (c *Container) BindSingleton(iface interface{}, impl interface{}) error {
	return c.BindWithOptions(iface, impl, true)
}

// BindWithOptions 带选项的绑定
func (c *Container) BindWithOptions(iface interface{}, impl interface{}, singleton bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	ifaceType := reflect.TypeOf(iface)
	if ifaceType.Kind() != reflect.Ptr {
		return fmt.Errorf("interface must be a pointer, got %v", ifaceType.Kind())
	}

	ifaceType = ifaceType.Elem()
	c.bindings[ifaceType] = binding{
		implementation: impl,
		singleton:      singleton,
	}

	return nil
}
