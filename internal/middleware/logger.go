package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// APILog 访问日志模型
type APILog struct {
	gorm.Model
	Method    string        `json:"method"`
	Path      string        `json:"path"`
	Status    int           `json:"status"`
	Duration  time.Duration `json:"duration"`
	IP        string        `json:"ip"`
	UserAgent string        `json:"user_agent"`
	RequestID string        `json:"request_id"`
}

func (APILog) Table() string {
	return "api_logs"
}

// Logger 中间件
func Logger(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		// 记录访问日志
		log := &APILog{
			Method:    c.Request.Method,
			Path:      path,
			Status:    c.Writer.Status(),
			Duration:  time.Since(start),
			IP:        c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			RequestID: c.GetString("RequestID"),
		}

		db.Create(log)
	}
}
