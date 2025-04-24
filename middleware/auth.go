package middleware

import (
	"net/http"
	"strings"

	"github.com/kruily/gofastcrud/fast_casbin"
	"github.com/kruily/gofastcrud/fast_jwt"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	jwtMaker    *fast_jwt.JWTMaker
	casbinMaker *fast_casbin.CasbinMaker
}

func NewAuthMiddleware(jwtMaker *fast_jwt.JWTMaker, casbinMaker *fast_casbin.CasbinMaker) *AuthMiddleware {
	return &AuthMiddleware{
		jwtMaker:    jwtMaker,
		casbinMaker: casbinMaker,
	}
}

// JWT 验证JWT token的中间件
func (m *AuthMiddleware) JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			return
		}

		fields := strings.Fields(authHeader)
		if len(fields) < 2 || fields[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		claims, err := m.jwtMaker.VerifyToken(fields[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		cla := claims.(*fast_jwt.Claims) // TODO: 需要根据实际情况修改Claims类型
		// 将用户信息存储到上下文中
		c.Set("user_id", cla.UserID)
		c.Set("username", cla.Username)

		c.Next()
	}
}

// Authorize 权限验证中间件
func (m *AuthMiddleware) Authorize(obj string, act string) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, exists := c.Get("username")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
			return
		}

		// 检查权限
		allowed, err := m.casbinMaker.Enforce(username.(string), obj, act)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to check permission"})
			return
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "permission denied"})
			return
		}

		c.Next()
	}
}
