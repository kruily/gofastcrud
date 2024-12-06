# GoFastCrud

GoFastCrud 是一个基于 Gin 框架的快速 CRUD 开发框架，帮助开发者快速构建 RESTful API。

## 特性

- 🚀 快速生成标准 CRUD 接口
- 📚 自动生成 Swagger 文档
- 🛠 支持自定义控制器和路由
- 🔌 灵活的中间件支持
- 🎯 类型安全的泛型实现
- 📦 工厂模式简化注册流程
- 💡 支持自定义响应处理

## 安装

```bash
go get github.com/kruily/GoFastCrud
```

## 使用

### 1. 配置

```go
// config.yaml
server: 
  address: ":8080"  // 服务地址

database:
  driver: "mysql" // 数据库驱动
  host: "localhost" // 数据库地址
  port: 3306 // 数据库端口
  username: "root" // 数据库用户名
  password: "password" // 数据库密码
  database: "test_crud" // 数据库名称
```
### 2. 启动服务
```go
// main.go
// 加载配置
cfg := config.Load("example/config/config.yaml")
// 创建数据库管理器
db := database.New()
if err := db.Init(cfg.Database); err != nil {
    log.Fatalf("Failed to initialize database: %v", err)
}
// 创建服务实例
srv := server.NewServer(cfg)
// 发布路由
srv.Publish("/api/v1")

// 运行服务（包含优雅启停）
if err := srv.Run(); err != nil {
    log.Fatalf("Server error: %v", err)
}
```

### 3. 定义实体模型
```go
go
// models/user.go
type User struct {
	ID        uint   `json:"id" gorm:"primarykey"`
	Username  string `json:"username" binding:"required" description:"用户名"`
	Email     string `json:"email" description:"邮箱地址"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
func (u User) SetID(id uint) {
	u.ID = id
}
```

### 4. 创建控制器
有两种方式创建控制器：

#### 4.1 使用标准控制器
这种方式会自动生成 CRUD 接口，并注册到路由中。
```go
// 创建控制器工厂
factory := crud.NewControllerFactory(db)
// 注册标准控制器(srv为服务实例)
crud.Register[*models.User](factory, "/users", srv)
```

#### 4.2 使用自定义控制器

```go
// controllers/user_controller.go
type UserController struct {
    // 嵌入 CrudController
    *crud.CrudController[models.User]
}
// 创建控制器实例
func NewUserController(db *gorm.DB) *UserController {
    controller := &UserController{
        CrudController: crud.NewCrudController(db, models.User{}),
    }
    // 应用中间件（可选）
    controller.UseMiddleware("*", middleware.Auth())
    // 某类方法应用中间件（可选）
    controller.UseMiddleware("POST", middleware.Validate())

    // 添加自定义路由
    controller.AddRoute(crud.APIRoute{
        Path:        "/login",
        Method:      "POST",
        // swagger 信息
        Tags:        []string{controller.GetEntityName()},
        Summary:     "用户登录",
        Description: "通过用户名和密码进行登录",
        // 请求处理函数
        Handler:     controller.Login,
        // 只对当前路由应用中间件（可选）
        Middleware:  []gin.HandlerFunc{middleware.Auth()},
    })
    return controller
}

// 注册自定义控制器
crud.RegisterCustomController[models.User](factory, "/users", srv, controllers.NewUserController)
```

### 5. 启用swagger
```go
// main.go
srv.EnableSwagger()
```

### 6. 完整代码
```go
// main.go
func main() {
    // 加载配置
    cfg := config.Load("config.yaml")

    // 初始化数据库
    db := database.NewDB()
    
    // 注册迁移模型
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
    srv := server.NewServer()
    srv.Publish("/api/v1")

    // 创建控制器工厂
    factory := crud.NewControllerFactory(db.DB())

    // 注册标准控制器
    crud.Register[*models.Book](factory, "/books", srv)

    // 标准控制器应用中间件
    c := crud.Register[*models.Phone](factory, "/phones", srv)
    c.UseMiddleware("*", middleware.Auth())

    // 注册自定义控制器
    crud.RegisterCustomController[models.User](
        factory,
        "/users",
        srv,
        controllers.NewUserController,
    )

    // 启用 Swagger
    srv.EnableSwagger()

    // 运行服务
    if err := srv.Run(); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}
```

## API 文档

启动服务后访问 `/swagger` 查看自动生成的 API 文档。

### 标准 CRUD 接口

- `GET /{entity}` - 获取列表
- `POST /{entity}` - 创建实体
- `GET /{entity}/{id}` - 获取单个实体
- `POST /{entity}/{id}` - 更新实体
- `DELETE /{entity}/{id}` - 删除实体

## 高级特性

### 中间件支持

```go
// 全局中间件
controller.UseMiddleware("*", middleware.Auth())

// 方法特定中间件
controller.UseMiddleware("POST", middleware.Validate())
```

### 自定义响应处理
```go
crud.SetConfig(&crud.CrudConfig{
    Responser: &CustomResponser{},
})
```
`CustomResponser` 需要实现 `ICrudResponse` 接口
```go
// internal/crud/response.go
type ICrudResponse interface {
	Success(data interface{}) interface{}
	Error(err error) interface{}
	List(items interface{}, total int64) interface{}
}
```

### 分页配置

```go
crud.SetConfig(&crud.CrudConfig{
    DefaultPageSize: 10,
    MaxPageSize:     100,
})
```

## 示例
查看 `example/` 目录获取完整示例。

## 贡献指南

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情