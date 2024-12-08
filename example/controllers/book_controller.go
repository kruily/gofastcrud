package controllers

import (
	"github.com/kruily/GoFastCrud/example/models"
	"github.com/kruily/GoFastCrud/internal/crud"
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
