package types

import (
	"reflect"
)

type RegisterServer interface {
	RegisterCrudController(path string, controller interface{}, entityType reflect.Type)
	RegisterCrudControllerWithFather(father any, path string, controller interface{}, entityType reflect.Type)
}
