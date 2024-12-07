package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode 错误码类型
type ErrorCode int

// AppError 应用错误
type AppError struct {
	Code    ErrorCode   `json:"code"`              // 错误码
	Message string      `json:"message"`           // 错误信息
	Details interface{} `json:"details,omitempty"` // 详细信息
	Err     error       `json:"-"`                 // 原始错误
	Stack   string      `json:"-"`                 // 错误堆栈
}

// 预定义错误码
const (
	// 系统级错误码 (1000-1999)
	ErrInternal     ErrorCode = 1000
	ErrUnauthorized ErrorCode = 1001
	ErrForbidden    ErrorCode = 1002
	ErrNotFound     ErrorCode = 1003
	ErrValidation   ErrorCode = 1004
	ErrTimeout      ErrorCode = 1005

	// 业务级错误码 (2000-2999)
	ErrUserNotFound    ErrorCode = 2000
	ErrUserExists      ErrorCode = 2001
	ErrInvalidPassword ErrorCode = 2002

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

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// WithDetails 添加详细信息
func (e *AppError) WithDetails(details interface{}) *AppError {
	e.Details = details
	return e
}

// WithError 添加原始错误
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// HTTPStatus 获取对应的HTTP状态码
func (e *AppError) HTTPStatus() int {
	if status, ok := httpStatusMap[e.Code]; ok {
		return status
	}
	return http.StatusInternalServerError
}

// New 创建新的应用错误
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Wrap 包装已有错误
func Wrap(err error, code ErrorCode, message string) *AppError {
	if err == nil {
		return nil
	}
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Is 检查错误类型
func Is(err error, code ErrorCode) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == code
	}
	return false
}
