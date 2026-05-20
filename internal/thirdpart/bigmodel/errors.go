package bigmodel

import "fmt"

// APIError API错误结构体
type APIError struct {
	Code    int    `json:"code"`              // HTTP状态码
	Message string `json:"message"`           // 错误消息
	Type    string `json:"type,omitempty"`    // 错误类型
	Param   string `json:"param,omitempty"`   // 相关参数
	Detail  string `json:"detail,omitempty"`  // 详细信息
}

// Error 实现error接口
func (e *APIError) Error() string {
	if e.Type != "" {
		return fmt.Sprintf("bigmodel API error [%d]: %s (type: %s)", e.Code, e.Message, e.Type)
	}
	return fmt.Sprintf("bigmodel API error [%d]: %s", e.Code, e.Message)
}

// IsAPIError 检查错误是否为API错误
func IsAPIError(err error) bool {
	_, ok := err.(*APIError)
	return ok
}

// GetAPIError 获取API错误详情
func GetAPIError(err error) *APIError {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr
	}
	return nil
}

// 预定义的错误类型
var (
	// ErrInvalidAPIKey API密钥无效
	ErrInvalidAPIKey = &APIError{
		Code:    401,
		Message: "Invalid API key",
		Type:    "authentication_error",
	}

	// ErrRateLimitExceeded 请求频率超限
	ErrRateLimitExceeded = &APIError{
		Code:    429,
		Message: "Rate limit exceeded",
		Type:    "rate_limit_error",
	}

	// ErrInsufficientQuota 配额不足
	ErrInsufficientQuota = &APIError{
		Code:    429,
		Message: "Insufficient quota",
		Type:    "quota_error",
	}

	// ErrInvalidRequest 请求参数无效
	ErrInvalidRequest = &APIError{
		Code:    400,
		Message: "Invalid request",
		Type:    "invalid_request_error",
	}

	// ErrModelNotFound 模型不存在
	ErrModelNotFound = &APIError{
		Code:    404,
		Message: "Model not found",
		Type:    "model_error",
	}

	// ErrServerError 服务器内部错误
	ErrServerError = &APIError{
		Code:    500,
		Message: "Internal server error",
		Type:    "server_error",
	}
)

// NewAPIError 创建新的API错误
func NewAPIError(code int, message, errorType string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Type:    errorType,
	}
}

// IsRetryableError 判断错误是否可重试
func IsRetryableError(err error) bool {
	apiErr := GetAPIError(err)
	if apiErr == nil {
		return false
	}

	// 5xx服务器错误和429限流错误可以重试
	return apiErr.Code >= 500 || apiErr.Code == 429
}

// IsAuthenticationError 判断是否为认证错误
func IsAuthenticationError(err error) bool {
	apiErr := GetAPIError(err)
	if apiErr == nil {
		return false
	}
	return apiErr.Code == 401
}

// IsRateLimitError 判断是否为限流错误
func IsRateLimitError(err error) bool {
	apiErr := GetAPIError(err)
	if apiErr == nil {
		return false
	}
	return apiErr.Code == 429
}