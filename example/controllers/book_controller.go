package controllers

import (
	"github.com/kruily/GoFastCrud/core/crud"
	"github.com/kruily/GoFastCrud/example/models"
	"gorm.io/gorm"
)

type BookController struct {
	*crud.CrudController[models.Book]
}

func NewBookController(db *gorm.DB) *BookController {
	controller := &BookController{
		CrudController: crud.NewCrudController(db, models.Book{}),
	}
	return controller
}
