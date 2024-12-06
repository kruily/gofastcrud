package server

import (
	"context"
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
	"github.com/kruily/GoFastCrud/internal/swagger"
)

type Server struct {
	config     *config.Config
	router     *gin.Engine
	srv        *http.Server
	apiGroup   *gin.RouterGroup
	swaggerGen *swagger.Generator
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
		config:     cfg,
		router:     r,
		srv:        srv,
		swaggerGen: swagger.NewGenerator(),
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
	swaggerGroup := s.router.Group("/swagger")
	{
		// 提供 Swagger UI HTML
		swaggerGroup.GET("/", func(c *gin.Context) {
			c.Header("Content-Type", "text/html")
			c.String(200, swaggerUITemplate, s.config.Server.Address)
		})

		// API 文档数据
		swaggerGroup.GET("/doc.json", func(c *gin.Context) {
			c.JSON(200, s.swaggerGen.GetAllSwagger())
		})
	}
}

// Swagger UI 模板
const swaggerUITemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Fast CRUD API</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css">
    <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist@5.11.0/favicon-32x32.png" sizes="32x32" />
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin: 0; background: #fafafa; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = () => {
            const ui = SwaggerUIBundle({
                url: "/swagger/doc.json",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                defaultModelsExpandDepth: 1,
                defaultModelExpandDepth: 1,
                defaultModelRendering: 'example',
                displayRequestDuration: true,
                docExpansion: 'list',
                filter: true,
                showExtensions: true
            });
            window.ui = ui;
        };
    </script>
</body>
</html>
`

// RegisterCrudController 注册 CRUD 控制器并生成文档
func (s *Server) RegisterCrudController(path string, controller interface{}, entityType reflect.Type) {
	routePath := strings.TrimPrefix(path, "/")

	if c, ok := controller.(interface{ RegisterRoutes(*gin.RouterGroup) }); ok {
		c.RegisterRoutes(s.apiGroup.Group(path))
	}

	// 传入 controller 以生成路由文档
	s.swaggerGen.RegisterEntity(entityType, s.apiGroup.BasePath(), routePath, controller)
}
