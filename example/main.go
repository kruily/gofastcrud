package main

import (
	"log"
	"os"

	"github.com/kruily/GoFastCrud/example/controllers"
	"github.com/kruily/GoFastCrud/example/models"
	"github.com/kruily/GoFastCrud/internal/config"
	"github.com/kruily/GoFastCrud/internal/crud"
	"github.com/kruily/GoFastCrud/internal/database"
	"github.com/kruily/GoFastCrud/internal/server"
	"github.com/kruily/GoFastCrud/pkg/utils"
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
	// 创建配置管理器
	configManager := config.NewConfigManager()

	// 加载配置
	configPath := os.Getenv("CONFIG_PATH")
	if err := configManager.LoadConfig(configPath); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 注册配置重载回调
	configManager.OnReload(func() {
		log.Println("Configuration reloaded")
		// 这里可以添加重载后的处理逻辑
	})

	cfg := configManager.GetConfig()

	// 创建数据库管理器
	db := database.New()

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
	// logConfig := logger.Config{
	// 	Level: logger.InfoLevel,
	// 	FileConfig: &logger.FileConfig{
	// 		Filename:   "logs/app.log",
	// 		MaxSize:    100,
	// 		MaxBackups: 3,
	// 		MaxAge:     7,
	// 		Compress:   true,
	// 	},
	// 	ConsoleLevel: logger.DebugLevel,
	// }

	// log, err := logger.NewLogger(logConfig)
	// if err != nil {
	// 	log.Fatal("Failed to create logger: %v", map[string]interface{}{"error": err})
	// }
	// defer log.Close()

	// 创建服务实例
	srv := server.NewServer(cfg)
	// 发布路由
	srv.PublishVersion(server.V1)
	srv.PublishVersion(server.V2)

	// 创建控制器工厂
	factory := crud.NewControllerFactory(db.DB())

	// 注册用户控制器
	crud.RegisterCustomController[*models.User](factory, "/users", srv, controllers.NewUserController)

	// 注册书籍控制器
	crud.Register[*models.Book](factory, "/books", srv)
	// 注册分类控制器
	crud.Register[*models.Category](factory, "/categories", srv)
	// 为控制器添加中间件
	// crud.Register[*models.Book](factory, "/books", srv).
	// 	UseMiddleware("*", gin.Logger())

	// 运行服务（包含优雅启停）
	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
