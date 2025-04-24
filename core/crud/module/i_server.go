package module

import (
	"reflect"

	"github.com/kruily/gofastcrud/core/crud/types"
)

type IServer interface {
	IModule
	PublishVersion(version types.APIVersion)
	Run() error
	RegisterCrudController(path string, controller interface{}, entityType reflect.Type)
	EnableSwagger()
}
