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

// RegisterBatch 批量注册实体
func (f *ControllerFactory) RegisterBatch(server types.RegisterServer, models ...ICrudEntity) {
	for _, model := range models {
		server.RegisterCrudController(model.Table(), NewCrudController(f.db, model), reflect.TypeOf(model))
	}
}

// RegisterBatchCustom 批量注册自定义控制器
func (f *ControllerFactory) RegisterBatchCustom(server types.RegisterServer, controllerConstructor ...func(*gorm.DB) ICrudController[ICrudEntity]) {
	for _, constructor := range controllerConstructor {
		controller := constructor(f.db)
		server.RegisterCrudController(controller.GetEntityName(), controller, reflect.TypeOf(controller.GetEntity()))
	}
}

// RegisterBatchMap 批量注册实体
// 使用场景：当需要批量注册实体,希望自定义路由组名，可以通过此方法注册
// server 是注册服务器
// models 是实体映射
func (f *ControllerFactory) RegisterBatchMap(server types.RegisterServer, models map[string]ICrudEntity) {
	for key, model := range models {
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
		server.RegisterCrudController(key, controller, reflect.TypeOf(controller.GetEntity()))
	}
}
