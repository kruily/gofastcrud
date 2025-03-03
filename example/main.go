package main

import (
	"log"
	"os"

	"github.com/kruily/gofastcrud/core/app"
	"github.com/kruily/gofastcrud/core/crud"
	"github.com/kruily/gofastcrud/core/server"
	"github.com/kruily/gofastcrud/utils"
)

// @title Fast CRUD API
// @version 1.0
// @description 自动生成的 CRUD API
// @BasePath /api/v1
func init() {
	// 获取项目根目录
	projectRoot, err := os.Getwd()
	if err != nil {
		log.Fatalf("无法获取项目根目录: %v", err)
	}
	// 加载环境变量
	utils.LoadEnv(projectRoot)
}

func main() {

	app := app.NewDefaultGoFastCrudApp()

	app.PublishVersion(server.V1)

	// app.RegisterModels(&models.User{}, &models.Book{}, &models.Category{})

	app.RegisterControllers(func(factory *crud.ControllerFactory, server *server.Server) {
		// 注册默认控制器
		// factory.Register(server, models.User{})
		// 使用自定义控制器
		// factory.RegisterCustom(server, controllers.NewUserController)

		// 使用批量注册
		// factory.RegisterBatch(server, &models.User{}, &models.Book{}, &models.Category{})

		// 使用批量注册自定义控制器
		// factory.RegisterBatchCustom(server, controllers.NewUserController, controllers.NewBookController)

		// 自定义路由名称（批量）
		// factory.RegisterBatchMap(server, map[string]crud.ICrudEntity{
		// 	"users": &models.User{},
		// 	"books": &models.Book{},
		// 	"categories": &models.Category{},
		// })

		// 自定义路由名称（批量自定义控制器）
		// factory.RegisterBatchCustomMap(server, map[string]func(*gorm.DB) crud.ICrudController[crud.ICrudEntity]{
		// 	"users": controllers.NewUserController,
		// 	"books": controllers.NewBookController,
		// 	"categories": controllers.NewCategoryController,
		// })

		// 路由继承
		// g := factory.Register(server, &models.User{})
		// _ = factory.RegisterWithFather(server, g, &models.Book{})

		// 路由继承自定义控制器
		// g := factory.RegisterCustom(server, controllers.NewUserController)
		// _ = factory.RegisterWithFatherCustom(server, g, controllers.NewBookController)
	})
	app.Start()
}
