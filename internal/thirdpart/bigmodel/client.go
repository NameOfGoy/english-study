package bigmodel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// BaseURL 智谱AI API基础地址
	BaseURL = "https://open.bigmodel.cn/api/paas/v4"
	// DefaultTimeout 默认超时时间
	DefaultTimeout = 30 * time.Second
)

// Client 智谱AI SDK客户端
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient 创建新的智谱AI客户端
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: BaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// NewClientWithConfig 使用自定义配置创建客户端
func NewClientWithConfig(apiKey, baseURL string, timeout time.Duration) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// SetTimeout 设置请求超时时间
func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// doRequest 执行HTTP请求的通用方法
func (c *Client) doRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := c.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", "bigmodel-go-sdk/1.0.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}

	return resp, nil
}

// handleResponse 处理HTTP响应的通用方法
func (c *Client) handleResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err != nil {
			return &APIError{
				Code:    resp.StatusCode,
				Message: string(body),
			}
		}
		apiErr.Code = resp.StatusCode
		return &apiErr
	}

	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}

	return nil
}

// handleBinaryResponse 处理二进制响应的方法（用于TTS等返回音频数据的API）
func (c *Client) handleBinaryResponse(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		// 尝试解析错误响应
		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err != nil {
			return nil, &APIError{
				Code:    resp.StatusCode,
				Message: string(body),
			}
		}
		apiErr.Code = resp.StatusCode
		return nil, &apiErr
	}

	return body, nil
}

// Chat 获取对话服务
func (c *Client) Chat() *ChatService {
	return NewChatService(c)
}

// Image 获取图像服务
func (c *Client) Image() *ImageService {
	return NewImageService(c)
}

// TTS 获取文本转语音服务
func (c *Client) TTS() *TTSService {
	return NewTTSService(c)
}
