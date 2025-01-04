package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/kruily/gofastcrud/errors"
	"github.com/kruily/gofastcrud/logger"
)

// ErrorHandler 错误处理中间件
func ErrorHandler(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// 记录panic堆栈信息
				stack := string(debug.Stack())
				log.Error("Panic recovered", map[string]interface{}{
					"error": r,
					"stack": stack,
				})

				appErr := &errors.AppError{
					Code:    errors.ErrInternal,
					Message: "Internal server error",
					Stack:   stack,
				}

				c.AbortWithStatusJSON(http.StatusInternalServerError, appErr)
			}
		}()

		c.Next()

		// 处理请求过程中设置的错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			var appErr *errors.AppError

			// 转换为应用错误
			switch e := err.(type) {
			case *errors.AppError:
				appErr = e
			default:
				appErr = errors.Wrap(err, errors.ErrInternal, "Internal server error")
			}

			// 记录错误日志
			log.Error(appErr.Message, map[string]interface{}{
				"code":    appErr.Code,
				"error":   appErr.Err,
				"details": appErr.Details,
				"path":    c.Request.URL.Path,
				"method":  c.Request.Method,
			})

			// 返回错误响应
			c.JSON(appErr.HTTPStatus(), appErr)
		}
	}
}
