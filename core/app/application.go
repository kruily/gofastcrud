package app

import (
	"log"

	"github.com/kruily/gofastcrud/config"
	"github.com/kruily/gofastcrud/core/crud"
	"github.com/kruily/gofastcrud/core/crud/module"
	"github.com/kruily/gofastcrud/core/crud/types"
	"github.com/kruily/gofastcrud/core/database"
	"github.com/kruily/gofastcrud/core/server"
	"github.com/kruily/gofastcrud/logger"
)

// gofastcrud Application 应用
type GoFastCrudApp struct {
	configManager *config.ConfigManager
	db            *database.Database
	server        *server.Server
	factory       *crud.ControllerFactory
	logger        *logger.Logger
}

// NewDefaultApplication 创建默认应用
func NewDefaultGoFastCrudApp() *GoFastCrudApp {
	// 获取配置管理器
	configManager := module.CRUD_MODULE.GetService(module.ConfigService).(*config.ConfigManager)
	configManager.LoadConfig()
	// 获取数据库模组
	db := module.CRUD_MODULE.GetService(module.DatabaseService).(*database.Database)
	// 创建服务
	server := server.NewServer(configManager.GetConfig())
	// 注册服务
	module.CRUD_MODULE.WithServer(server)

	// 创建控制器工厂
	factory := crud.NewControllerFactory(db.DB())

	return &GoFastCrudApp{
		configManager: configManager,
		db:            db,
		server:        server,
		factory:       factory,
	}
}

// RegisterModels 注册模型
func (a *GoFastCrudApp) RegisterModels(models ...interface{}) {
	a.db.RegisterModels(models...)
}

// RegisterControllers 注册控制器
func (a *GoFastCrudApp) RegisterControllers(fn func(*crud.ControllerFactory, *server.Server)) {
	fn(a.factory, a.server)
}

func (a *GoFastCrudApp) PublishVersion(version types.APIVersion) {
	a.server.PublishVersion(version)
}

func (a *GoFastCrudApp) GetServer() *server.Server {
	return a.server
}

func (a *GoFastCrudApp) GetLogger() *logger.Logger {
	return a.logger
}

func (a *GoFastCrudApp) WithLogger(logger *logger.Logger) {
	a.logger = logger
	module.CRUD_MODULE.WithLogger(logger)
}

// Start 启动应用
func (a *GoFastCrudApp) Start() {
	a.db.Init(&a.configManager.GetConfig().Database)
	// 运行服务（包含优雅启停）
	if err := a.server.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
