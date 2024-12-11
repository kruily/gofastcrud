package crud

import (
	"reflect"

	"github.com/kruily/gofastcrud/core/crud/types"
	"gorm.io/gorm"
)

// ControllerFactory 控制器工厂
type ControllerFactory struct {
	db *gorm.DB
}

// NewControllerFactory 创建控制器工厂
func NewControllerFactory(db *gorm.DB) *ControllerFactory {
	return &ControllerFactory{db: db}
}

// Register 注册实体并创建控制器
// 使用场景：当需要使用默认的控制器时，可以通过此方法注册
// f 是控制器工厂
// path 是路由路径
// server 是注册服务器
func Register[T ICrudEntity](
	f *ControllerFactory,
	path string,
	server types.RegisterServer,
) *CrudController[T] {
	var entity T
	controller := NewCrudController(f.db, entity)
	server.RegisterCrudController(path, controller, reflect.TypeOf(entity))
	return controller
}

func RegisterWithOptions[T ICrudEntity](
	f *ControllerFactory,
	path string,
	server types.RegisterServer,
	opts ...func(*CrudController[T]),
) *CrudController[T] {
	controller := Register[T](f, path, server)
	for _, opt := range opts {
		opt(controller)
	}
	return controller
}

// RegisterCustomController 注册自定义控制器
// 使用场景：当需要使用自定义的控制器时，可以通过此方法注册
// 自定义控制器需要实现 CrudControllerInterface 接口
// f 是控制器工厂
// path 是路由路径
// server 是注册服务器
// controllerConstructor 是控制器构造函数
func RegisterCustomController[T ICrudEntity](
	f *ControllerFactory,
	path string,
	server types.RegisterServer,
	controllerConstructor func(*gorm.DB) ICrudController[T],
) ICrudController[T] {
	controller := controllerConstructor(f.db)
	server.RegisterCrudController(path, controller, reflect.TypeOf(*new(T)))
	return controller
}
