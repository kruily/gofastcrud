package swagger

// SwaggerDoc 自定义路由的 Swagger 文档
type SwaggerDoc struct {
	Path        string      // 路径
	Method      string      // HTTP 方法
	Summary     string      // 摘要
	Description string      // 描述
	Tags        []string    // 标签
	Request     interface{} // 请求体结构
	Response    interface{} // 响应体结构
}

// SwaggerRouter Swagger 文档接口
type SwaggerRouter interface {
	SwaggerDocs() []SwaggerDoc
}
