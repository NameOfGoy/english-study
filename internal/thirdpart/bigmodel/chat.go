package bigmodel

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	// ChatCompletionsEndpoint 对话补全API端点
	ChatCompletionsEndpoint = "/chat/completions"
)

// ChatService 对话服务
type ChatService struct {
	client *Client
}

// NewChatService 创建对话服务
func NewChatService(client *Client) *ChatService {
	return &ChatService{
		client: client,
	}
}

// CreateChatCompletion 创建对话补全（非流式）
func (s *ChatService) CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	// 确保非流式请求
	req.Stream = false

	// 验证请求参数
	if err := s.validateChatRequest(&req); err != nil {
		return nil, err
	}

	// 发送请求
	resp, err := s.client.doRequest(ctx, "POST", ChatCompletionsEndpoint, req)
	if err != nil {
		return nil, fmt.Errorf("chat completion request failed: %w", err)
	}

	// 解析响应
	var result ChatCompletionResponse
	if err := s.client.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateChatCompletionStream 创建流式对话补全
func (s *ChatService) CreateChatCompletionStream(ctx context.Context, req ChatCompletionRequest, callback StreamCallback) error {
	// 确保流式请求
	req.Stream = true

	// 验证请求参数
	if err := s.validateChatRequest(&req); err != nil {
		return err
	}

	if callback == nil {
		return fmt.Errorf("stream callback is required")
	}

	// 发送请求
	resp, err := s.client.doRequest(ctx, "POST", ChatCompletionsEndpoint, req)
	if err != nil {
		return fmt.Errorf("chat completion stream request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode >= 400 {
		return s.client.handleResponse(resp, nil)
	}

	// 处理流式响应
	return s.handleStreamResponse(resp, callback)
}

// handleStreamResponse 处理流式响应
func (s *ChatService) handleStreamResponse(resp *http.Response, callback StreamCallback) error {
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行
		if line == "" {
			continue
		}

		// 检查是否为SSE数据行
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		// 提取数据部分
		data := strings.TrimPrefix(line, "data: ")

		// 检查是否为结束标记
		if data == "[DONE]" {
			break
		}

		// 解析JSON数据
		var streamResp ChatCompletionStreamResponse
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			// 忽略解析错误，继续处理下一行
			continue
		}

		// 调用回调函数
		if err := callback(streamResp); err != nil {
			return fmt.Errorf("stream callback error: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("stream reading error: %w", err)
	}

	return nil
}

// validateChatRequest 验证对话请求参数
func (s *ChatService) validateChatRequest(req *ChatCompletionRequest) error {
	if req.Model == "" {
		return fmt.Errorf("model is required")
	}

	if len(req.Messages) == 0 {
		return fmt.Errorf("messages are required")
	}

	// 验证消息格式
	for i, msg := range req.Messages {
		if msg.Role == "" {
			return fmt.Errorf("message[%d]: role is required", i)
		}
		if msg.Content == "" {
			return fmt.Errorf("message[%d]: content is required", i)
		}
		if !isValidRole(msg.Role) {
			return fmt.Errorf("message[%d]: invalid role '%s'", i, msg.Role)
		}
	}

	// 验证温度参数
	if req.Temperature != nil && (*req.Temperature < 0 || *req.Temperature > 1) {
		return fmt.Errorf("temperature must be between 0 and 1")
	}

	// 验证top_p参数
	if req.TopP != nil && (*req.TopP < 0 || *req.TopP > 1) {
		return fmt.Errorf("top_p must be between 0 and 1")
	}

	// 验证max_tokens参数
	if req.MaxTokens != nil && *req.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be positive")
	}

	return nil
}

// isValidRole 检查角色是否有效
func isValidRole(role string) bool {
	validRoles := []string{"system", "user", "assistant"}
	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	return false
}

// CreateSimpleChat 创建简单对话（便捷方法）
func (s *ChatService) CreateSimpleChat(ctx context.Context, model, userMessage string) (*ChatCompletionResponse, error) {
	req := ChatCompletionRequest{
		Model: model,
		Messages: []Message{
			{
				Role:    "user",
				Content: userMessage,
			},
		},
	}

	return s.CreateChatCompletion(ctx, req)
}

// CreateChatWithSystem 创建带系统提示的对话（便捷方法）
func (s *ChatService) CreateChatWithSystem(ctx context.Context, model, systemMessage, userMessage string) (*ChatCompletionResponse, error) {
	req := ChatCompletionRequest{
		Model: model,
		Messages: []Message{
			{
				Role:    "system",
				Content: systemMessage,
			},
			{
				Role:    "user",
				Content: userMessage,
			},
		},
	}

	return s.CreateChatCompletion(ctx, req)
}

// GetChatResponse 获取对话响应内容（便捷方法）
func GetChatResponse(resp *ChatCompletionResponse) string {
	if resp == nil || len(resp.Choices) == 0 {
		return ""
	}
	return resp.Choices[0].Message.Content
}