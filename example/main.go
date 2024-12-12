package main

import (
	"log"
	"os"

	"github.com/kruily/gofastcrud/core/crud"
	"github.com/kruily/gofastcrud/core/crud/module"
	"github.com/kruily/gofastcrud/core/database"
	"github.com/kruily/gofastcrud/core/server"
	"github.com/kruily/gofastcrud/example/controllers"
	"github.com/kruily/gofastcrud/example/models"
	"github.com/kruily/gofastcrud/pkg/config"
	"github.com/kruily/gofastcrud/pkg/logger"
	"github.com/kruily/gofastcrud/pkg/utils"
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
	configManager := module.CRUD_MODULE.GetService(module.ConfigService).(*config.ConfigManager)
	configManager.LoadConfig()
	// 注册配置重载回调
	configManager.OnReload(func() {
		log.Println("Configuration reloaded")
		// 这里可以添加重载后的处理逻辑
	})
	// 获取配置
	cfg := configManager.GetConfig()

	// 获取数据库模组
	db := module.CRUD_MODULE.GetService(module.DatabaseService).(*database.Database)

	// 注册模型
	db.RegisterModels(
		&models.User{},
		&models.Book{},
		&models.Category{},
		// 添加其他模型
	)

	// 初始化数据库
	if err := db.Init(&cfg.Database); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 创建日志实例
	logService, err := logger.NewLogger(logger.Config{
		Level: logger.InfoLevel,
		FileConfig: &logger.FileConfig{
			Filename:   "logs/app.log",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   true,
		},
		ConsoleLevel: logger.DebugLevel,
	})
	if err != nil {
		log.Fatal("Failed to create logger: ", err)
	}
	defer logService.Close()

	// 注册日志服务
	module.CRUD_MODULE.WithLogger(logService)

	// 创建服务实例
	srv := server.NewServer(cfg)
	// 注册服务
	module.CRUD_MODULE.WithServer(srv)
	// 发布路由
	srv.PublishVersion(server.V1)
	srv.PublishVersion(server.V2)

	// 创建控制器工厂
	factory := crud.NewControllerFactory(db.DB())
	factory.RegisterBatchCustom(srv, controllers.NewUserController, controllers.NewBookController)
	factory.RegisterBatch(srv, &models.Category{})

	// 运行服务（包含优雅启停）
	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
