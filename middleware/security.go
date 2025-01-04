package middleware

import (
	"html"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SecurityHeaders 安全头中间件
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 安全响应头
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Next()
	}
}

// XSSProtection XSS防护中间件
func XSSProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 处理请求参数
		for _, param := range c.Params {
			c.Params[0].Value = html.EscapeString(param.Value)
		}

		// 处理查询参数
		for key, values := range c.Request.URL.Query() {
			for i, value := range values {
				values[i] = html.EscapeString(value)
			}
			c.Request.URL.Query()[key] = values
		}

		// 处理POST表单数据
		if c.Request.Method == http.MethodPost {
			if err := c.Request.ParseForm(); err == nil {
				for key, values := range c.Request.PostForm {
					for i, value := range values {
						values[i] = html.EscapeString(value)
					}
					c.Request.PostForm[key] = values
				}
			}
		}

		c.Next()
	}
}

// SQLInjectionProtection SQL注入防护中间件
func SQLInjectionProtection() gin.HandlerFunc {
	// SQL注入特征正则表达式
	sqlInjectionPattern := regexp.MustCompile(`(?i)(SELECT|INSERT|UPDATE|DELETE|DROP|UNION|ALTER|CREATE|WHERE|OR|AND)\s`)

	return func(c *gin.Context) {
		// 检查查询参数
		for _, values := range c.Request.URL.Query() {
			for _, value := range values {
				if sqlInjectionPattern.MatchString(value) {
					c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
						"error": "Potential SQL injection detected",
					})
					return
				}
			}
		}

		// 检查POST表单数据
		if c.Request.Method == http.MethodPost {
			if err := c.Request.ParseForm(); err == nil {
				for _, values := range c.Request.PostForm {
					for _, value := range values {
						if sqlInjectionPattern.MatchString(value) {
							c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
								"error": "Potential SQL injection detected",
							})
							return
						}
					}
				}
			}
		}

		c.Next()
	}
}

// CORSConfig CORS配置
func CORSConfig() gin.HandlerFunc {
	config := cors.Config{
		AllowOrigins:     []string{"*"},                                       // 允许的源
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH"},   // 允许的HTTP方法
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"}, // 允许的头
		ExposeHeaders:    []string{"Content-Length"},                          // 暴露的头
		AllowCredentials: true,                                                // 允许携带凭证
		MaxAge:           12 * time.Hour,                                      // 预检请求结果缓存时间
	}
	return cors.New(config)
}
