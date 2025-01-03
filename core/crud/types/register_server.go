package types

import (
	"reflect"

	"github.com/gin-gonic/gin"
)

type RegisterServer interface {
	RegisterCrudController(path string, controller interface{}, entityType reflect.Type)
	RegisterWithGroup(group *gin.RouterGroup, path string, controller interface{}, entityType reflect.Type)
}
