package controllers

import (
	"github.com/kruily/gofastcrud/core/crud"
	"github.com/kruily/gofastcrud/example/models"
	"gorm.io/gorm"
)

type BookController struct {
	*crud.CrudController[models.Book]
}

func NewBookController(db *gorm.DB) crud.ICrudController[crud.ICrudEntity] {
	controller := &BookController{
		CrudController: crud.NewCrudController[models.Book](db, models.Book{}),
	}
	return controller
}
