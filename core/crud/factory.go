package crud

import (
	"reflect"

	"github.com/kruily/gofastcrud/core/crud/types"
	"gorm.io/gorm"
)

// ControllerFactory 控制器工厂
type ControllerFactory struct {
	db     *gorm.DB
	models []interface{}
}

// NewControllerFactory 创建控制器工厂
func NewControllerFactory(db *gorm.DB) *ControllerFactory {
	return &ControllerFactory{db: db}
}

// RegisterGroup 注册后台路由组
// 简单注册，默认注册
// server 是注册服务器
// model 是实体
func (f *ControllerFactory) Register(server types.RegisterServer, model ICrudEntity) ICrudController[ICrudEntity] {
	f.models = append(f.models, model)
	controller := NewCrudController(f.db, model)
	server.RegisterCrudController(model.Table(), controller, reflect.TypeOf(model))
	return controller
}

// RegisterCustom 注册自定义控制器
// 使用场景：当需要注册自定义控制器，可以通过此方法注册
// server 是注册服务器
// model 是实体
// constructor 是控制器构造函数
func (f *ControllerFactory) RegisterCustom(server types.RegisterServer, constructor func(*gorm.DB) ICrudController[ICrudEntity]) ICrudController[ICrudEntity] {
	controller := constructor(f.db)
	f.models = append(f.models, controller.GetEntity())
	server.RegisterCrudController(controller.GetEntity().Table(), controller, reflect.TypeOf(controller.GetEntity()))
	return controller
}

// RegisterWithGroup 注册后台路由组
// 使用场景：当需要注册后台路由组，并且此路由在另外的路由之下，可以通过此方法注册
// server 是注册服务器
// group 是路由组
// model 是实体
func (f *ControllerFactory) RegisterWithFather(server types.RegisterServer, father ICrudController[ICrudEntity], model ICrudEntity) ICrudController[ICrudEntity] {
	f.models = append(f.models, model)
	controller := NewCrudController(f.db, model)
	server.RegisterCrudControllerWithFather(father, model.Table(), controller, reflect.TypeOf(model))
	return controller
}

func (f *ControllerFactory) RegisterWithFatherCustom(server types.RegisterServer, father ICrudController[ICrudEntity], constructor func(*gorm.DB) ICrudController[ICrudEntity]) ICrudController[ICrudEntity] {
	controller := constructor(f.db)
	f.models = append(f.models, controller.GetEntity())
	server.RegisterCrudControllerWithFather(father, controller.GetEntity().Table(), controller, reflect.TypeOf(controller.GetEntity()))
	return controller
}

// RegisterBatch 批量注册实体
// 使用场景：当需要批量注册实体,且没有自定义控制器，可以通过此方法注册
// server 是注册服务器
// models 是实体
func (f *ControllerFactory) RegisterBatch(server types.RegisterServer, models ...ICrudEntity) {
	for _, model := range models {
		f.models = append(f.models, model)
		server.RegisterCrudController(model.Table(), NewCrudController(f.db, model), reflect.TypeOf(model))
	}
}

// RegisterBatchCustom 批量注册自定义控制器
// 使用场景：当需要批量注册自定义控制器,可以通过此方法注册
// server 是注册服务器
// controllerConstructor 是控制器构造函数
func (f *ControllerFactory) RegisterBatchCustom(server types.RegisterServer, controllerConstructor ...func(*gorm.DB) ICrudController[ICrudEntity]) {
	for _, constructor := range controllerConstructor {
		controller := constructor(f.db)
		f.models = append(f.models, controller.GetEntity())
		server.RegisterCrudController(controller.GetEntity().Table(), controller, reflect.TypeOf(controller.GetEntity()))
	}
}

// RegisterBatchMap 批量注册实体
// 使用场景：当需要批量注册实体,希望自定义路由组名，可以通过此方法注册
// server 是注册服务器
// models 是实体映射
func (f *ControllerFactory) RegisterBatchMap(server types.RegisterServer, models map[string]ICrudEntity) {
	for key, model := range models {
		f.models = append(f.models, model)
		server.RegisterCrudController(key, NewCrudController(f.db, model), reflect.TypeOf(model))
	}
}

// RegisterBatchCustomMap 批量注册自定义控制器
// 使用场景：当需要批量注册自定义控制器,希望自定义路由组名，可以通过此方法注册
// server 是注册服务器
// mapControllerConstructor 是控制器构造函数映射
func (f *ControllerFactory) RegisterBatchCustomMap(server types.RegisterServer, mapControllerConstructor map[string]func(*gorm.DB) ICrudController[ICrudEntity]) {
	for key, constructor := range mapControllerConstructor {
		controller := constructor(f.db)
		f.models = append(f.models, controller.GetEntity())
		server.RegisterCrudController(key, controller, reflect.TypeOf(controller.GetEntity()))
	}
}

func (f *ControllerFactory) Migrate() {
	f.db.AutoMigrate(f.models...)
}
