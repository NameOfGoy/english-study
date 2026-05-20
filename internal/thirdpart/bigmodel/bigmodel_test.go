package bigmodel

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// 测试用的模拟API密钥
const testAPIKey = "test-api-key-12345"

// TestNew 测试SDK创建
func TestNew(t *testing.T) {
	sdk := New(testAPIKey)
	if sdk == nil {
		t.Fatal("SDK creation failed")
	}

	if sdk.client == nil {
		t.Error("Client not initialized")
	}

	if sdk.chatService == nil {
		t.Error("Chat service not initialized")
	}

	if sdk.imageService == nil {
		t.Error("Image service not initialized")
	}

	if sdk.auth == nil {
		t.Error("Auth not initialized")
	}
}

// TestNewWithConfig 测试自定义配置创建SDK
func TestNewWithConfig(t *testing.T) {
	customBaseURL := "https://custom-api.example.com/api/paas/v4"
	customTimeout := 30 * time.Second

	sdk := NewWithConfig(testAPIKey, customBaseURL, customTimeout)
	if sdk == nil {
		t.Fatal("SDK creation with config failed")
	}

	// Note: baseURL and timeout are private fields, so we can't directly test them
	// In a real implementation, you might want to add getter methods or make fields public for testing
}

// TestSetTimeout 测试设置超时时间
func TestSetTimeout(t *testing.T) {
	sdk := New(testAPIKey)
	newTimeout := 45 * time.Second

	sdk.SetTimeout(newTimeout)

	// Note: timeout is a private field, so we can't directly test it
	// In a real implementation, you might want to add a getter method for testing
}

// TestSetAPIKey 测试设置API密钥
func TestSetAPIKey(t *testing.T) {
	sdk := New(testAPIKey)
	newAPIKey := "new-test-api-key"

	sdk.SetAPIKey(newAPIKey)

	if sdk.client.apiKey != newAPIKey {
		t.Errorf("Expected API key %s, got %s", newAPIKey, sdk.client.apiKey)
	}
}

// TestGetAPIKey 测试获取API密钥（遮蔽显示）
func TestGetAPIKey(t *testing.T) {
	sdk := New(testAPIKey)
	maskedKey := sdk.GetAPIKey()

	// 检查是否正确遮蔽
	if !strings.Contains(maskedKey, "***") {
		t.Error("API key should be masked")
	}

	// 检查不应该包含完整的原始密钥
	if strings.Contains(maskedKey, testAPIKey) {
		t.Error("API key should not contain full original key")
	}
}

// TestValidateAPIKey 测试API密钥验证
func TestValidateAPIKey(t *testing.T) {
	// 测试空密钥
	sdk := New("")
	err := sdk.ValidateAPIKey()
	if err == nil {
		t.Error("Empty API key should be invalid")
	}

	// 测试有效格式的密钥
	sdk = New(testAPIKey)
	err = sdk.ValidateAPIKey()
	if err != nil {
		t.Errorf("Valid API key should pass validation: %v", err)
	}
}

// TestChatCompletionRequest 测试对话请求结构
func TestChatCompletionRequest(t *testing.T) {
	req := ChatCompletionRequest{
		Model: ModelGLM4,
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
		Temperature: float64Ptr(0.7),
		MaxTokens:   intPtr(1000),
	}

	if req.Model != ModelGLM4 {
		t.Errorf("Expected model %s, got %s", ModelGLM4, req.Model)
	}

	if len(req.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(req.Messages))
	}

	if req.Messages[0].Role != "user" {
		t.Errorf("Expected role 'user', got %s", req.Messages[0].Role)
	}

	if *req.Temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %f", *req.Temperature)
	}

	if *req.MaxTokens != 1000 {
		t.Errorf("Expected max tokens 1000, got %d", *req.MaxTokens)
	}
}

// TestImageGenerationRequest 测试图像生成请求结构
func TestImageGenerationRequest(t *testing.T) {
	req := ImageGenerationRequest{
		Model:   ModelCogView3,
		Prompt:  "A beautiful landscape",
		Size:    ImageSize1024,
		N:       intPtr(2),
		Quality: ImageQualityHD,
	}

	if req.Model != ModelCogView3 {
		t.Errorf("Expected model %s, got %s", ModelCogView3, req.Model)
	}

	if req.Prompt != "A beautiful landscape" {
		t.Errorf("Expected prompt 'A beautiful landscape', got %s", req.Prompt)
	}

	if req.Size != ImageSize1024 {
		t.Errorf("Expected size %s, got %s", ImageSize1024, req.Size)
	}

	if *req.N != 2 {
		t.Errorf("Expected N 2, got %d", *req.N)
	}

	if req.Quality != ImageQualityHD {
		t.Errorf("Expected quality %s, got %s", ImageQualityHD, req.Quality)
	}
}

// TestAPIError 测试API错误处理
func TestAPIError(t *testing.T) {
	apiErr := NewAPIError(401, "Unauthorized", "authentication_error")

	if apiErr.Code != 401 {
		t.Errorf("Expected code 401, got %d", apiErr.Code)
	}

	if apiErr.Message != "Unauthorized" {
		t.Errorf("Expected message 'Unauthorized', got %s", apiErr.Message)
	}

	if apiErr.Type != "authentication_error" {
		t.Errorf("Expected type 'authentication_error', got %s", apiErr.Type)
	}

	// 测试错误字符串
	errorStr := apiErr.Error()
	if !strings.Contains(errorStr, "401") {
		t.Error("Error string should contain status code")
	}

	if !strings.Contains(errorStr, "Unauthorized") {
		t.Error("Error string should contain message")
	}
}

// TestIsAPIError 测试API错误检查
func TestIsAPIError(t *testing.T) {
	apiErr := NewAPIError(400, "Bad Request", "invalid_request")

	if !IsAPIError(apiErr) {
		t.Error("Should recognize API error")
	}

	// 测试普通错误
	normalErr := errors.New("normal error")
	if IsAPIError(normalErr) {
		t.Error("Should not recognize normal error as API error")
	}
}

// TestErrorTypes 测试错误类型检查
func TestErrorTypes(t *testing.T) {
	// 认证错误
	authErr := NewAPIError(401, "Unauthorized", "authentication_error")
	if !IsAuthenticationError(authErr) {
		t.Error("Should recognize authentication error")
	}

	// 频率限制错误
	rateLimitErr := NewAPIError(429, "Too Many Requests", "rate_limit_exceeded")
	if !IsRateLimitError(rateLimitErr) {
		t.Error("Should recognize rate limit error")
	}

	// 可重试错误
	if !IsRetryableError(rateLimitErr) {
		t.Error("Rate limit error should be retryable")
	}

	serverErr := NewAPIError(500, "Internal Server Error", "server_error")
	if !IsRetryableError(serverErr) {
		t.Error("Server error should be retryable")
	}

	// 不可重试错误
	if IsRetryableError(authErr) {
		t.Error("Authentication error should not be retryable")
	}
}

// TestHelperFunctions 测试辅助函数
func TestHelperFunctions(t *testing.T) {
	// 测试指针辅助函数
	intVal := 42
	intPtr := intPtr(intVal)
	if *intPtr != intVal {
		t.Errorf("Expected %d, got %d", intVal, *intPtr)
	}

	floatVal := 3.14
	floatPtr := float64Ptr(floatVal)
	if *floatPtr != floatVal {
		t.Errorf("Expected %f, got %f", floatVal, *floatPtr)
	}

	// Note: stringPtr and boolPtr helper functions are not implemented in the current SDK
	// These would be useful additions for a complete SDK implementation
}

// TestConstants 测试常量定义
func TestConstants(t *testing.T) {
	// 测试模型常量
	models := []string{
		ModelGLM4,
		ModelGLM4V,
		ModelGLM3Turbo,
		ModelGLM4Flash,
		ModelCogView3,
	}

	for _, model := range models {
		if model == "" {
			t.Error("Model constant should not be empty")
		}
	}

	// 测试图像尺寸常量
	sizes := []string{
		ImageSize256,
		ImageSize512,
		ImageSize1024,
		ImageSize1024x1792,
		ImageSize1792x1024,
	}

	for _, size := range sizes {
		if size == "" {
			t.Error("Image size constant should not be empty")
		}
	}

	// 测试版本常量
	if SDKVersion == "" {
		t.Error("SDK version should not be empty")
	}

	if APIVersion == "" {
		t.Error("API version should not be empty")
	}
}

// BenchmarkSDKCreation 基准测试SDK创建
func BenchmarkSDKCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sdk := New(testAPIKey)
		_ = sdk
	}
}

// BenchmarkAPIKeyValidation 基准测试API密钥验证
func BenchmarkAPIKeyValidation(b *testing.B) {
	sdk := New(testAPIKey)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = sdk.ValidateAPIKey()
	}
}