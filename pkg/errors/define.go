package errors

import "net/http"

// ErrorCode 错误码类型
type ErrorCode int

// 预定义错误码
const (
	// 系统级错误码 (1000-1999)
	ErrInternal     ErrorCode = 1000
	ErrUnauthorized ErrorCode = 1001
	ErrForbidden    ErrorCode = 1002
	ErrNotFound     ErrorCode = 1003
	ErrValidation   ErrorCode = 1004
	ErrTimeout      ErrorCode = 1005
	ErrIDType       ErrorCode = 1006

	// 业务级错误码 (2000-2999)
	ErrUserNotFound    ErrorCode = 2000
	ErrUserExists      ErrorCode = 2001
	ErrInvalidPassword ErrorCode = 2002
	ErrInvalidParam    ErrorCode = 2003

	// 数据库错误码 (3000-3999)
	ErrDatabase       ErrorCode = 3000
	ErrDuplicateKey   ErrorCode = 3001
	ErrNoRowsAffected ErrorCode = 3002

	// 第三方服务错误码 (4000-4999)
	ErrThirdParty ErrorCode = 4000
	ErrRateLimit  ErrorCode = 4001
)

// 错误码与HTTP状态码的映射
var httpStatusMap = map[ErrorCode]int{
	ErrInternal:        http.StatusInternalServerError,
	ErrUnauthorized:    http.StatusUnauthorized,
	ErrForbidden:       http.StatusForbidden,
	ErrNotFound:        http.StatusNotFound,
	ErrValidation:      http.StatusBadRequest,
	ErrTimeout:         http.StatusGatewayTimeout,
	ErrUserNotFound:    http.StatusNotFound,
	ErrUserExists:      http.StatusConflict,
	ErrInvalidPassword: http.StatusBadRequest,
	ErrDatabase:        http.StatusInternalServerError,
	ErrDuplicateKey:    http.StatusConflict,
	ErrNoRowsAffected:  http.StatusNotFound,
	ErrThirdParty:      http.StatusBadGateway,
	ErrRateLimit:       http.StatusTooManyRequests,
}

func RegisterErrorCode(code ErrorCode, message string, httpStatus int) error {
	if _, exists := httpStatusMap[code]; exists {
		return New(ErrInternal, "Error code already exists")
	}
	httpStatusMap[code] = httpStatus
	return nil
}
