package main

import (
	"log"

	"github.com/kruily/GoFastCrud/example/controllers"
	"github.com/kruily/GoFastCrud/example/models"
	"github.com/kruily/GoFastCrud/internal/config"
	"github.com/kruily/GoFastCrud/internal/crud"
	"github.com/kruily/GoFastCrud/internal/database"
	"github.com/kruily/GoFastCrud/internal/server"
)

// @title Fast CRUD API
// @version 1.0
// @description 自动生成的 CRUD API
// @BasePath /api/v1
func main() {
	// 加载配置
	cfg := config.Load("example/config/config.yaml")

	// 创建数据库管理器
	db := database.New()

	// 注册模型
	db.RegisterModels(
		&models.User{},
		&models.Book{},
		// 添加其他模型
	)

	// 初始化数据库
	if err := db.Init(cfg.Database); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 创建服务实例
	srv := server.NewServer(cfg)
	// 发布路由
	srv.Publish("/api/v1")

	// 创建控制器工厂
	factory := crud.NewControllerFactory(db.DB())

	// 注册用户控制器
	crud.RegisterCustomController[*models.User](factory, "/users", srv, controllers.NewUserController)

	// 注册书籍控制器
	crud.Register[*models.Book](factory, "/books", srv)
	// 为控制器添加中间件
	// crud.Register[*models.Book](factory, "/books", srv).
	// 	UseMiddleware("*", gin.Logger())

	// 运行服务（包含优雅启停）
	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
