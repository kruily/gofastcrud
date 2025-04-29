package app

import (
	"log"

	"github.com/kruily/gofastcrud/config"
	"github.com/kruily/gofastcrud/core/crud"
	"github.com/kruily/gofastcrud/core/crud/module"
	"github.com/kruily/gofastcrud/core/crud/types"
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

	container := di.SINGLE()

	opt := &AppOption{}
	for _, o := range opts {
		o(opt)
	}
	// 创建服务
	server := server.NewServer(config.CONFIG_MANAGER.GetConfig())
	// 创建控制器工厂
	factory := crud.NewControllerFactory()

	if _, err := container.ResolveSingleton(module.ResponseService); err == nil {
		container.BindSingletonWithName(module.ResponseService, &utils.DefaultResponseHandler{})
	}

	return &GoFastCrudApp{
		server:    server,
		factory:   factory,
		container: container,
	}
}

// RegisterControllers 注册控制器
func (a *GoFastCrudApp) RegisterControllers(fn func(*crud.ControllerFactory, *server.Server)) *GoFastCrudApp {
	fn(a.factory, a.server)
	a.factory.Migrate() // 自动迁移数据库
	return a
}

func (a *GoFastCrudApp) PublishVersion(version types.APIVersion) *GoFastCrudApp {
	a.server.PublishVersion(version)
	return a
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
