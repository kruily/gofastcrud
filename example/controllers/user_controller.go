package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/kruily/GoFastCrud/example/models"
	"github.com/kruily/GoFastCrud/internal/crud"
	"github.com/kruily/GoFastCrud/internal/crud/types"
)

var validate = validator.New()

// CreateRequest 创建用户请求
type CreateRequest struct {
	Username string `json:"username" validate:"required,min=3,max=32"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type UserController struct {
	*crud.CrudController[models.User]
}

func NewUserController(db *gorm.DB) *UserController {
	controller := &UserController{
		CrudController: crud.NewCrudController(db, models.User{}),
	}
	controller.AddRoute(types.APIRoute{
		Method:  http.MethodPost,
		Path:    "/register",
		Summary: "注册用户",
		Tags:    []string{controller.GetEntityName()},
		Handler: controller.Create,
	})
	return controller
}

func (c *UserController) Create(ctx *gin.Context) (interface{}, error) {
	var request CreateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request parameters",
			"error":   err.Error(),
		})
		return nil, err
	}

	if err := validate.Struct(request); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "Internal validation error",
				"error":   err.Error(),
			})
			return nil, err
		}

		var errors []string
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, err.Error())
		}
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Validation failed",
			"errors":  errors,
		})
		return nil, err
	}

	// 在这里处理创建用户的逻辑
	return nil, nil
}
