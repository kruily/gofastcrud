package crud

import (
	"reflect"

	"github.com/kruily/gofastcrud/core/crud/types"
	"gorm.io/gorm"
)

// ControllerFactory 控制器工厂
type ControllerFactory[T ID_TYPE] struct {
	db *gorm.DB
}

// NewControllerFactory 创建控制器工厂
func NewControllerFactory[T ID_TYPE](db *gorm.DB) *ControllerFactory[T] {
	return &ControllerFactory[T]{db: db}
}

// RegisterBatch 批量注册实体
func (f *ControllerFactory[T]) RegisterBatch(server types.RegisterServer, models ...ICrudEntity[T]) {
	for _, model := range models {
		server.RegisterCrudController(model.Table(), NewCrudController(f.db, model), reflect.TypeOf(model))
	}
}

// RegisterBatchCustom 批量注册自定义控制器
func (f *ControllerFactory[T]) RegisterBatchCustom(server types.RegisterServer, controllerConstructor ...func(*gorm.DB) ICrudController[ICrudEntity[T], T]) {
	for _, constructor := range controllerConstructor {
		controller := constructor(f.db)
		server.RegisterCrudController(controller.GetEntityName(), controller, reflect.TypeOf(controller.GetEntity()))
	}
}

// RegisterBatchMap 批量注册实体
// 使用场景：当需要批量注册实体,希望自定义路由组名，可以通过此方法注册
// server 是注册服务器
// models 是实体映射
func (f *ControllerFactory[T]) RegisterBatchMap(server types.RegisterServer, models map[string]ICrudEntity[T]) {
	for key, model := range models {
		server.RegisterCrudController(key, NewCrudController(f.db, model), reflect.TypeOf(model))
	}
}

// RegisterBatchCustomMap 批量注册自定义控制器
// 使用场景：当需要批量注册自定义控制器,希望自定义路由组名，可以通过此方法注册
// server 是注册服务器
// mapControllerConstructor 是控制器构造函数映射
func (f *ControllerFactory[T]) RegisterBatchCustomMap(server types.RegisterServer, mapControllerConstructor map[string]func(*gorm.DB) ICrudController[ICrudEntity[T], T]) {
	for key, constructor := range mapControllerConstructor {
		controller := constructor(f.db)
		server.RegisterCrudController(key, controller, reflect.TypeOf(controller.GetEntity()))
	}
}
