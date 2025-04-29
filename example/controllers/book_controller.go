package controllers

import (
	"github.com/kruily/gofastcrud/core/crud"
	"github.com/kruily/gofastcrud/core/database"
	"github.com/kruily/gofastcrud/example/models"
)

type BookController struct {
	*crud.CrudController[*models.Book]
}

func NewBookController(db *database.Database) crud.ICrudController[crud.ICrudEntity] {
	controller := &BookController{
		CrudController: crud.NewCrudController(db, &models.Book{}),
	}
	return controller
}
