package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kruily/GoFastCrud/internal/config"
	"github.com/kruily/GoFastCrud/internal/swagger" // swagger files
)

type Server struct {
	config        *config.Config
	router        *gin.Engine
	srv           *http.Server
	apiGroup      *gin.RouterGroup
	swaggerGen    *swagger.Generator
	enableSwagger bool
}

// RouteRegister 路由注册函数类型
type RouteRegister func(r *gin.RouterGroup)

func NewServer(cfg *config.Config) *Server {
	// 创建 Gin 引擎
	r := gin.Default()

	// 创建 HTTP 服务
	srv := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: r,
	}

	return &Server{
		config:        cfg,
		router:        r,
		srv:           srv,
		swaggerGen:    swagger.NewGenerator(),
		enableSwagger: cfg.Server.EnableSwagger,
	}
}

// Publish 设置 API 路径前缀
func (s *Server) Publish(apiPath string) *Server {
	s.apiGroup = s.router.Group(apiPath)
	return s
}

// RegisterRoutes 注册路由
func (s *Server) RegisterRoutes(register RouteRegister) {
	if s.apiGroup == nil {
		s.apiGroup = s.router.Group("/") // 默认使用根路径
	}
	register(s.apiGroup)
}

// Run 启动服务并处理优雅关闭
func (s *Server) Run() error {
	// 启用 Swagger 文档
	if s.enableSwagger {
		s.EnableSwagger()
	}

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
	// 创建自定义的 swagger handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/swagger/doc.json" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(s.swaggerGen.GetAllSwagger())
			return
		}
		swagger.SwaggerUIHandler(w, r)
	})

	// 注册 Swagger UI 路由
	s.router.GET("/swagger/*any", gin.WrapH(handler))
}

// RegisterCrudController 注册 CRUD 控制器并生成文档
func (s *Server) RegisterCrudController(path string, controller interface{}, entityType reflect.Type) {
	routePath := strings.TrimPrefix(path, "/")

	if c, ok := controller.(interface{ RegisterRoutes(*gin.RouterGroup) }); ok {
		c.RegisterRoutes(s.apiGroup.Group(path))
	}

	// 传入 controller 以生成路由文档
	if s.enableSwagger {
		s.swaggerGen.RegisterEntity(entityType, s.apiGroup.BasePath(), routePath, controller)
	}
}
