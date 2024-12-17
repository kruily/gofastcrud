package controllers

import (
	"github.com/google/uuid"
	"github.com/kruily/gofastcrud/core/crud"
	"github.com/kruily/gofastcrud/example/models"
	"gorm.io/gorm"
)

type BookController struct {
	*crud.CrudController[models.Book, uuid.UUID]
}

func NewBookController(db *gorm.DB) crud.ICrudController[crud.ICrudEntity[uuid.UUID], uuid.UUID] {
	controller := &BookController{
		CrudController: crud.NewCrudController[models.Book, uuid.UUID](db, models.Book{}),
	}
	return controller
}
