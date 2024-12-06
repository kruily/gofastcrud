package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/kruily/GoFastCrud/example/models"
	"github.com/kruily/GoFastCrud/internal/crud"
	"github.com/kruily/GoFastCrud/internal/crud/types"
	"gorm.io/gorm"
)

// UserController 用户控制器
type UserController struct {
	*crud.CrudController[models.User]
}

// NewUserController 创建用户控制器s
func NewUserController(db *gorm.DB) *UserController {
	controller := &UserController{
		CrudController: crud.NewCrudController(db, models.User{}),
	}
	// 整个控制器添加中间件
	controller.UseMiddleware("*", gin.Logger())
	// 控制器的某个HTTP方法添加中间件
	controller.UseMiddleware("POST", gin.Logger())

	// 添加自定义路由
	controller.AddRoute(types.APIRoute{
		Path:        "/login",
		Method:      "POST",
		Tags:        []string{controller.GetEntityName()},
		Summary:     "用户登录",
		Description: "通过用户名和密码进行登录",
		Handler:     controller.Login,
		Request:     LoginRequest{},
		Response:    LoginResponse{},
	})

	controller.AddRoute(types.APIRoute{
		Path:        "/profile",
		Method:      "GET",
		Tags:        []string{controller.GetEntityName()},
		Summary:     "获取用户资料",
		Description: "获取当前登录用户的详细信息",
		Handler:     controller.GetProfile,
		// 单独添加中间件
		Middlewares: []gin.HandlerFunc{gin.Logger()},
	})

	return controller
}

// Login 用户登录
func (c *UserController) Login(ctx *gin.Context) (interface{}, error) {
	// 登录逻辑实现
	return gin.H{"message": "login success"}, nil
}

// GetProfile 获取用户资料
func (c *UserController) GetProfile(ctx *gin.Context) (interface{}, error) {
	// 获取用户资料逻辑实现
	return gin.H{"message": "get profile success"}, nil
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" example:"john_doe" description:"用户名"`
	Password string `json:"password" example:"password123" description:"密码"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJ..." description:"JWT令牌"`
}
