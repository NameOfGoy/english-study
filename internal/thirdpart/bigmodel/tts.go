package bigmodel

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// TTSEndpoint TTS API端点
	TTSEndpoint = "/audio/speech"
	// DefaultTTSModel 默认TTS模型
	DefaultTTSModel = "cogtts"
)

// TTSService TTS服务
type TTSService struct {
	client *Client
}

// NewTTSService 创建TTS服务
func NewTTSService(client *Client) *TTSService {
	return &TTSService{
		client: client,
	}
}

// CreateTTS 创建TTS请求
func (s *TTSService) CreateTTS(ctx context.Context, req TTSRequest) (*TTSResponse, error) {
	// 设置默认模型
	if req.Model == "" {
		req.Model = DefaultTTSModel
	}

	// 设置默认音色
	if req.Voice == "" {
		req.Voice = string(TTSVoiceTongtong)
	}

	// 验证请求参数
	if err := s.validateTTSRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// 发送请求
	resp, err := s.client.doRequest(ctx, "POST", TTSEndpoint, req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// 处理二进制响应
	audioData, err := s.client.handleBinaryResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("handle response failed: %w", err)
	}

	// 生成文件名
	filename := s.generateFilename(req)

	return &TTSResponse{
		Data:     audioData,
		Filename: filename,
	}, nil
}

// SaveToFile 将TTS响应保存到文件
func (s *TTSService) SaveToFile(response *TTSResponse, filepath string) error {
	if response == nil {
		return fmt.Errorf("response is nil")
	}

	// 创建目录（如果不存在）
	dir := filepath[:strings.LastIndex(filepath, "/")]
	if dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// 写入文件
	if err := os.WriteFile(filepath, response.Data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// generateFilename 生成建议的文件名
func (s *TTSService) generateFilename(req TTSRequest) string {
	// 获取当前时间戳
	timestamp := time.Now().Format("20060102_150405")

	// 获取文件扩展名
	ext := ".wav" // 默认扩展名
	if req.ResponseFormat != nil {
		switch *req.ResponseFormat {
		case "mp3":
			ext = ".mp3"
		case "wav":
			ext = ".wav"
		}
	}

	// 生成文件名
	return fmt.Sprintf("tts_%s_%s%s", req.Voice, timestamp, ext)
}

// validateTTSRequest 验证TTS请求参数
func (s *TTSService) validateTTSRequest(req TTSRequest) error {
	if req.Model == "" {
		return fmt.Errorf("model is required")
	}

	if req.Input == "" {
		return fmt.Errorf("input text is required")
	}

	// 验证文本长度
	if len(req.Input) > 4096 {
		return fmt.Errorf("input text exceeds maximum length of 4096 characters")
	}

	if req.Voice == "" {
		return fmt.Errorf("voice is required")
	}

	// 验证音色
	validVoices := []string{"tongtong", "chuichui", "xiaochen", "jam", "kazi", "douji", "luodo"}
	if !contains(validVoices, req.Voice) {
		return fmt.Errorf("invalid voice: %s, supported voices: %v", req.Voice, validVoices)
	}

	// 验证语速
	if req.Speed != nil && (*req.Speed < 0.5 || *req.Speed > 2.0) {
		return fmt.Errorf("speed must be between 0.5 and 2.0")
	}

	// 验证音量
	if req.Volume != nil && (*req.Volume <= 0 || *req.Volume > 10) {
		return fmt.Errorf("volume must be between 0 and 10")
	}

	// 验证响应格式
	if req.ResponseFormat != nil {
		validFormats := []string{"wav", "mp3"}
		if !contains(validFormats, *req.ResponseFormat) {
			return fmt.Errorf("invalid response format: %s, supported formats: %v", *req.ResponseFormat, validFormats)
		}
	}

	return nil
}

// CreateTTSWithConfig 使用完整配置创建TTS（便捷方法）
func (s *TTSService) CreateTTSWithConfig(ctx context.Context, text string, voice TTSVoice, speed, volume *float64, format *TTSResponseFormat) (*TTSResponse, error) {
	req := TTSRequest{
		Model:  DefaultTTSModel,
		Input:  text,
		Voice:  string(voice),
		Speed:  speed,
		Volume: volume,
	}

	if format != nil {
		formatStr := string(*format)
		req.ResponseFormat = &formatStr
	}

	return s.CreateTTS(ctx, req)
}

// CreateSimpleTTS 创建简单的TTS请求（最简便的方法）
func (s *TTSService) CreateSimpleTTS(ctx context.Context, text string) (*TTSResponse, error) {
	return s.CreateTTSWithVoice(ctx, text, TTSVoiceTongtong)
}

// SaveAudioToFile 将TTS响应保存到指定文件（便捷方法）
func SaveAudioToFile(response *TTSResponse, filename string) error {
	if response == nil {
		return fmt.Errorf("response is nil")
	}

	// 创建目录（如果不存在）
	dir := filepath.Dir(filename)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// 写入文件
	if err := os.WriteFile(filename, response.Data, 0644); err != nil {
		return fmt.Errorf("failed to write audio file: %w", err)
	}

	return nil
}

// CreateTTSWithVoice 使用指定音色创建TTS（便捷方法）
func (s *TTSService) CreateTTSWithVoice(ctx context.Context, text string, voice TTSVoice) (*TTSResponse, error) {
	req := TTSRequest{
		Model: DefaultTTSModel,
		Input: text,
		Voice: string(voice),
	}

	return s.CreateTTS(ctx, req)
}



// 辅助函数
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsInt(slice []int, item int) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}