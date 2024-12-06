package utils

// Response 基础响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// ListResponse 列表响应结构
type ListResponse struct {
	Items interface{} `json:"items"`
	Total int64       `json:"total"`
}

// DefaultResponseHandler 默认响应处理器
type DefaultResponseHandler struct{}

// Success 处理成功响应
func (h *DefaultResponseHandler) Success(data interface{}) interface{} {
	return Response{
		Code:    200,
		Message: "success",
		Data:    data,
	}
}

// Error 处理错误响应
func (h *DefaultResponseHandler) Error(err error) interface{} {
	return Response{
		Code:    500,
		Message: err.Error(),
		Data:    nil,
	}
}

// List 处理列表响应
func (h *DefaultResponseHandler) List(items interface{}, total int64) interface{} {
	return Response{
		Code:    200,
		Message: "success",
		Data: ListResponse{
			Items: items,
			Total: total,
		},
	}
}
