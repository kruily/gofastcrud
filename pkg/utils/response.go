package utils

// Response 基础响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// PagenationResponse 分页响应结构
type PagenationResponse struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
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

// Pagenation 处理列表响应
func (h *DefaultResponseHandler) Pagenation(items interface{}, total int64, page int, size int) interface{} {
	return Response{
		Code:    200,
		Message: "success",
		Data: PagenationResponse{
			List:  items,
			Total: total,
			Page:  page,
			Size:  size,
		},
	}
}
