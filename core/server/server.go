package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kruily/gofastcrud/core/swagger" // swagger files
	"github.com/kruily/gofastcrud/core/templates"
	"github.com/kruily/gofastcrud/pkg/config"
	"github.com/kruily/gofastcrud/pkg/logger"
)

type Server struct {
	config         *config.Config
	router         *gin.Engine
	srv            *http.Server
	log            *logger.Logger
	swaggerGen     *swagger.Generator
	enableSwagger  bool
	versionManager *VersionManager
	apiGroups      map[APIVersion]*gin.RouterGroup
}

// RouteRegister 路由注册函数类型
type RouteRegister func(r *gin.RouterGroup)

func NewServer(cfg *config.Config) *Server {
	// 创建 Gin 引擎
	r := gin.Default()

	// 拼接地址
	address := fmt.Sprintf("%s:%d", cfg.Server.Address, cfg.Server.Port)

	// 创建 HTTP 服务
	srv := &http.Server{
		Addr:    address,
		Handler: r,
	}

	return &Server{
		config:         cfg,
		router:         r,
		srv:            srv,
		swaggerGen:     swagger.NewGenerator(),
		enableSwagger:  cfg.Server.EnableSwagger,
		versionManager: NewVersionManager(),
		apiGroups:      make(map[APIVersion]*gin.RouterGroup),
	}
}

// SetLogger 设置日志实例
func (s *Server) SetLogger(log *logger.Logger) {
	s.log = log
}

// Publish 设置 API 路径前缀
func (s *Server) PublishVersion(version APIVersion) *Server {
	if !s.versionManager.IsValidVersion(version) {
		s.versionManager.RegisterVersion(version)
	}
	path := fmt.Sprintf("/api/%s", version)
	group := s.router.Group(path)
	s.apiGroups[version] = group
	return s
}

// RegisterRoutes 注册路由
func (s *Server) RegisterRoutes(register RouteRegister) {
	for _, group := range s.apiGroups {
		register(group)
	}
}

// Run 启动服务并处理优雅关闭
func (s *Server) Run() error {
	// 启用 Swagger 文档
	if s.enableSwagger {
		s.EnableSwagger()
	}

	// 获取所有可用的API版本
	versions := s.versionManager.GetAvailableVersions()
	versionStrs := make([]string, len(versions))
	for i, v := range versions {
		versionStrs[i] = string(v)
	}

	// 注册主页路由
	s.router.GET("/", gin.WrapH(templates.HomeHandler(versionStrs)))

	// 启动服务
	go func() {
		log.Printf("Server starting on %s\n", s.config.Server.Address)
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server stopping...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		return err
	}

	log.Println("Server exiting")
	return nil
}

// Router 获取路由实例
func (s *Server) Router() *gin.Engine {
	return s.router
}

// EnableSwagger 启用 Swagger 文档
func (s *Server) EnableSwagger() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentVersion := s.versionManager.ParseVersionFromPath(r.URL.Path)
		versions := s.versionManager.GetAvailableVersions()
		versionStrs := make([]string, len(versions))
		for i, v := range versions {
			versionStrs[i] = string(v)
		}

		// 获取所有文档
		docs := s.swaggerGen.GetAllSwagger()

		swagger.SwaggerUIHandler(w, r, versionStrs, string(currentVersion), docs)
	})

	// 注册 Swagger UI 路由
	for version := range s.apiGroups {
		path := fmt.Sprintf("/api/%s/swagger/*any", version)
		s.router.GET(path, gin.WrapH(handler))
	}
}

// RegisterCrudController 注册 CRUD 控制器并生成文档
func (s *Server) RegisterCrudController(path string, controller interface{}, entityType reflect.Type) {
	for version, group := range s.apiGroups {
		routePath := strings.TrimPrefix(path, "/")
		if c, ok := controller.(interface{ RegisterRoutes(*gin.RouterGroup) }); ok {
			versionGroup := group.Group(path)
			c.RegisterRoutes(versionGroup)
		}

		// 生成对应版本的文档
		if s.enableSwagger {
			s.swaggerGen.RegisterEntityWithVersion(entityType, group.BasePath(), routePath, controller, string(version))
		}
	}
}
