package app

import (
	"log"

	"github.com/kruily/gofastcrud/config"
	"github.com/kruily/gofastcrud/core/crud"
	"github.com/kruily/gofastcrud/core/crud/types"
	"github.com/kruily/gofastcrud/core/database"
	"github.com/kruily/gofastcrud/core/di"
	"github.com/kruily/gofastcrud/core/server"
	"github.com/kruily/gofastcrud/logger"
	"github.com/kruily/gofastcrud/utils"
)

// gofastcrud Application 应用
type GoFastCrudApp struct {
	server    *server.Server
	factory   *crud.ControllerFactory
	logger    *logger.Logger
	container *di.Container
}

// NewDefaultApplication 创建默认应用
func NewDefaultGoFastCrudApp(opts ...Option) *GoFastCrudApp {
	opt := &AppOption{}
	for _, o := range opts {
		o(opt)
	}
	container := di.SINGLE()
	// 获取配置管理器
	configManager := config.NewConfigManager()
	container.BindSingletonWithName("CONFIG_MANAGER", configManager)
	configManager.LoadConfig()
	// 获取数据库模组
	db := database.New()
	db.Init(&configManager.GetConfig().Database)
	container.BindSingletonWithName("DATABASE", db)
	// 创建服务
	server := server.NewServer(configManager.GetConfig())

	// 创建控制器工厂
	factory := crud.NewControllerFactory(db.DB())

	if opt.Response != nil {
		container.BindSingletonWithName("RESPONSE_HANDLER", opt.Response)
	} else {
		container.BindSingletonWithName("RESPONSE_HANDLER", &utils.DefaultResponseHandler{})
	}

	return &GoFastCrudApp{
		server:    server,
		factory:   factory,
		container: container,
	}
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

// Start 启动应用
func (a *GoFastCrudApp) Start() {
	// 运行服务（包含优雅启停）
	if err := a.server.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
