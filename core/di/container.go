package di

import (
	"fmt"
	"reflect"
	"sync"
)

var singleContainer *Container

// Container 依赖注入容器
type Container struct {
	mu sync.RWMutex

	// 单例
	instances map[string]any
	// 绑定
	bindings map[string]binding
}

// binding 绑定信息
type binding struct {
	implementations []any
	singleton       bool
}

func NewBinding() binding {
	return binding{
		implementations: make([]any, 0),
		singleton:       false,
	}
}

func (b binding) AddImplementation(impl any) {
	b.implementations = append(b.implementations, impl)
}

func (b binding) IsSingleton() bool {
	return b.singleton
}

func (b binding) GetImplementations() []any {
	return b.implementations
}

// NewSingleContainer 创建单例容器
func SINGLE() *Container {
	if singleContainer == nil {
		singleContainer = New()
	}
	return singleContainer
}

// New 创建容器实例
func New() *Container {
	return &Container{
		instances: make(map[string]any),
		bindings:  make(map[string]binding),
	}
}

// BindSingleton 绑定单例
func (c *Container) BindSingleton(obj any) error {
	objType := reflect.TypeOf(obj)
	if objType.Kind() != reflect.Ptr {
		return fmt.Errorf("interface must be a pointer, got %v", objType.Kind())
	}

	objType = objType.Elem()
	return c.BindWithOptions(objType.Name(), obj, objType, true)
}

// BindSingletonWithName 绑定单例
func (c *Container) BindSingletonWithName(name string, obj any) error {
	objType := reflect.TypeOf(obj)
	if objType.Kind() != reflect.Ptr {
		return fmt.Errorf("interface must be a pointer, got %v", objType.Kind())
	}

	objType = objType.Elem()
	if name == "" {
		name = objType.Name()
	}
	return c.BindWithOptions(name, obj, objType, true)
}

// BindWithOptions 带选项的绑定
func (c *Container) BindWithOptions(name string, obj any, objType reflect.Type, singleton bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.instances[name] == nil && singleton {
		c.instances[name] = obj
		return nil
	}

	if _, ok := c.bindings[name]; !ok {
		c.bindings[name] = NewBinding()
	}
	binding := c.bindings[name]
	binding.AddImplementation(obj)
	binding.singleton = singleton
	c.bindings[name] = binding

	return nil
}
func (c *Container) ResolveSingleton(name string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if instance, ok := c.instances[name]; ok {
		return instance, nil
	}

	return nil, fmt.Errorf("no singleton instance found for type %v", name)
}

// 获取依赖
func (c *Container) GetSingletonByName(name string) (interface{}, error) {
	return c.ResolveSingleton(name)
}

// MustGetSingletonByName 获取依赖（如果出错则panic）
func (c *Container) MustGetSingletonByName(name string) interface{} {
	if result, err := c.GetSingletonByName(name); err != nil {
		panic(err)
	} else {
		return result
	}
}

func (c *Container) GetSingletonByType(obj any) (interface{}, error) {
	objType := reflect.TypeOf(obj)
	if objType.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("interface must be a pointer, got %v", objType.Kind())
	}

	objType = objType.Elem()
	return c.GetSingletonByName(objType.Name())
}

func (c *Container) MustGetSingletonByType(obj any) interface{} {
	if result, err := c.GetSingletonByType(obj); err != nil {
		panic(err)
	} else {
		return result
	}
}

func (c *Container) resolveImplementations(name string) ([]any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if bindings, ok := c.bindings[name]; ok {
		return bindings.GetImplementations(), nil
	}

	return nil, fmt.Errorf("no implementations found for type %v", name)
}

func (c *Container) GetImplementationsByName(name string) ([]any, error) {
	return c.resolveImplementations(name)
}

func (c *Container) MustGetImplementationsByName(name string) []any {
	if result, err := c.GetImplementationsByName(name); err != nil {
		panic(err)
	} else {
		return result
	}
}

func (c *Container) GetImplementationsByType(obj any) ([]any, error) {
	objType := reflect.TypeOf(obj)
	if objType.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("interface must be a pointer, got %v", objType.Kind())
	}

	objType = objType.Elem()
	return c.GetImplementationsByName(objType.Name())
}

func (c *Container) MustGetImplementationsByType(obj any) []any {
	if result, err := c.GetImplementationsByType(obj); err != nil {
		panic(err)
	} else {
		return result
	}
}
