package bigmodel

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// AuthConfig 认证配置
type AuthConfig struct {
	APIKey    string        // API密钥
	Timeout   time.Duration // 认证超时时间
	RetryMax  int           // 最大重试次数
	RetryWait time.Duration // 重试等待时间
}

// DefaultAuthConfig 默认认证配置
func DefaultAuthConfig(apiKey string) *AuthConfig {
	return &AuthConfig{
		APIKey:    apiKey,
		Timeout:   30 * time.Second,
		RetryMax:  3,
		RetryWait: 1 * time.Second,
	}
}

// Authenticator 认证器接口
type Authenticator interface {
	// Authenticate 为请求添加认证信息
	Authenticate(req *http.Request) error
	// ValidateAPIKey 验证API密钥格式
	ValidateAPIKey() error
	// GetAPIKey 获取API密钥
	GetAPIKey() string
}

// BearerTokenAuth Bearer Token认证器
type BearerTokenAuth struct {
	config *AuthConfig
}

// NewBearerTokenAuth 创建Bearer Token认证器
func NewBearerTokenAuth(apiKey string) *BearerTokenAuth {
	return &BearerTokenAuth{
		config: DefaultAuthConfig(apiKey),
	}
}

// NewBearerTokenAuthWithConfig 使用自定义配置创建Bearer Token认证器
func NewBearerTokenAuthWithConfig(config *AuthConfig) *BearerTokenAuth {
	return &BearerTokenAuth{
		config: config,
	}
}

// Authenticate 为请求添加Bearer Token认证头
func (auth *BearerTokenAuth) Authenticate(req *http.Request) error {
	if err := auth.ValidateAPIKey(); err != nil {
		return err
	}

	// 设置Authorization头
	req.Header.Set("Authorization", "Bearer "+auth.config.APIKey)
	return nil
}

// ValidateAPIKey 验证API密钥格式
func (auth *BearerTokenAuth) ValidateAPIKey() error {
	if auth.config.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	// 检查API密钥格式（智谱AI的API密钥通常以特定前缀开头）
	apiKey := strings.TrimSpace(auth.config.APIKey)
	if len(apiKey) < 10 {
		return fmt.Errorf("API key is too short")
	}

	// 检查是否包含非法字符
	for _, char := range apiKey {
		if char < 32 || char > 126 {
			return fmt.Errorf("API key contains invalid characters")
		}
	}

	return nil
}

// GetAPIKey 获取API密钥
func (auth *BearerTokenAuth) GetAPIKey() string {
	return auth.config.APIKey
}

// SetAPIKey 设置API密钥
func (auth *BearerTokenAuth) SetAPIKey(apiKey string) {
	auth.config.APIKey = apiKey
}

// SetTimeout 设置认证超时时间
func (auth *BearerTokenAuth) SetTimeout(timeout time.Duration) {
	auth.config.Timeout = timeout
}

// SetRetryConfig 设置重试配置
func (auth *BearerTokenAuth) SetRetryConfig(maxRetries int, retryWait time.Duration) {
	auth.config.RetryMax = maxRetries
	auth.config.RetryWait = retryWait
}

// AuthenticatedRequest 执行带认证的请求
func (auth *BearerTokenAuth) AuthenticatedRequest(ctx context.Context, client *http.Client, req *http.Request) (*http.Response, error) {
	// 添加认证信息
	if err := auth.Authenticate(req); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(ctx, auth.config.Timeout)
	defer cancel()
	req = req.WithContext(ctx)

	// 执行请求，支持重试
	var lastErr error
	for attempt := 0; attempt <= auth.config.RetryMax; attempt++ {
		resp, err := client.Do(req)
		if err == nil {
			// 检查是否为认证错误
			if resp.StatusCode == 401 {
				resp.Body.Close()
				return nil, ErrInvalidAPIKey
			}
			return resp, nil
		}

		lastErr = err

		// 如果不是最后一次尝试，等待后重试
		if attempt < auth.config.RetryMax {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(auth.config.RetryWait):
				// 继续重试
			}
		}
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", auth.config.RetryMax+1, lastErr)
}

// IsValidAPIKeyFormat 检查API密钥格式是否有效
func IsValidAPIKeyFormat(apiKey string) bool {
	auth := NewBearerTokenAuth(apiKey)
	return auth.ValidateAPIKey() == nil
}

// MaskAPIKey 遮蔽API密钥用于日志记录
func MaskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "****"
	}
	return apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
}