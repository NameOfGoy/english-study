package bigmodel

import (
	"context"
	"fmt"
	"strings"
)

const (
	// ImageGenerationEndpoint 图像生成API端点
	ImageGenerationEndpoint = "/images/generations"
)

// ImageService 图像服务
type ImageService struct {
	client *Client
}

// NewImageService 创建图像服务
func NewImageService(client *Client) *ImageService {
	return &ImageService{
		client: client,
	}
}

// CreateImage 创建图像
func (s *ImageService) CreateImage(ctx context.Context, req ImageGenerationRequest) (*ImageGenerationResponse, error) {
	// 验证请求参数
	if err := s.validateImageRequest(&req); err != nil {
		return nil, err
	}

	// 发送请求
	resp, err := s.client.doRequest(ctx, "POST", ImageGenerationEndpoint, req)
	if err != nil {
		return nil, fmt.Errorf("image generation request failed: %w", err)
	}

	// 解析响应
	var result ImageGenerationResponse
	if err := s.client.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// validateImageRequest 验证图像生成请求参数
func (s *ImageService) validateImageRequest(req *ImageGenerationRequest) error {
	if req.Model == "" {
		return fmt.Errorf("model is required")
	}

	if req.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}

	// 验证提示词长度
	if len(strings.TrimSpace(req.Prompt)) < 3 {
		return fmt.Errorf("prompt is too short")
	}

	if len(req.Prompt) > 4000 {
		return fmt.Errorf("prompt is too long (max 4000 characters)")
	}

	// 验证图像尺寸
	if req.Size != "" && !isValidImageSize(req.Size) {
		return fmt.Errorf("invalid image size: %s", req.Size)
	}

	// 验证生成数量
	if req.N != nil && (*req.N < 1 || *req.N > 10) {
		return fmt.Errorf("number of images must be between 1 and 10")
	}

	// 验证质量参数
	if req.Quality != "" && !isValidImageQuality(req.Quality) {
		return fmt.Errorf("invalid image quality: %s", req.Quality)
	}

	return nil
}

// isValidImageSize 检查图像尺寸是否有效
func isValidImageSize(size string) bool {
	validSizes := []string{
		"256x256",
		"512x512",
		"1024x1024",
		"1024x1792",
		"1792x1024",
	}

	for _, validSize := range validSizes {
		if size == validSize {
			return true
		}
	}
	return false
}

// isValidImageQuality 检查图像质量是否有效
func isValidImageQuality(quality string) bool {
	validQualities := []string{"standard", "hd"}

	for _, validQuality := range validQualities {
		if quality == validQuality {
			return true
		}
	}
	return false
}

// CreateSimpleImage 创建简单图像（便捷方法）
func (s *ImageService) CreateSimpleImage(ctx context.Context, model, prompt string) (*ImageGenerationResponse, error) {
	req := ImageGenerationRequest{
		Model:  model,
		Prompt: prompt,
		Size:   "1024x1024", // 默认尺寸
		N:      intPtr(1),    // 生成1张图片
	}

	return s.CreateImage(ctx, req)
}

// CreateImageWithSize 创建指定尺寸的图像（便捷方法）
func (s *ImageService) CreateImageWithSize(ctx context.Context, model, prompt, size string) (*ImageGenerationResponse, error) {
	req := ImageGenerationRequest{
		Model:  model,
		Prompt: prompt,
		Size:   size,
		N:      intPtr(1),
	}

	return s.CreateImage(ctx, req)
}

// CreateMultipleImages 创建多张图像（便捷方法）
func (s *ImageService) CreateMultipleImages(ctx context.Context, model, prompt string, count int) (*ImageGenerationResponse, error) {
	if count < 1 || count > 10 {
		return nil, fmt.Errorf("count must be between 1 and 10")
	}

	req := ImageGenerationRequest{
		Model:  model,
		Prompt: prompt,
		Size:   "1024x1024",
		N:      &count,
	}

	return s.CreateImage(ctx, req)
}

// CreateHDImage 创建高清图像（便捷方法）
func (s *ImageService) CreateHDImage(ctx context.Context, model, prompt string) (*ImageGenerationResponse, error) {
	req := ImageGenerationRequest{
		Model:   model,
		Prompt:  prompt,
		Size:    "1024x1024",
		N:       intPtr(1),
		Quality: "hd",
	}

	return s.CreateImage(ctx, req)
}

// GetImageURLs 获取生成的图像URL列表（便捷方法）
func GetImageURLs(resp *ImageGenerationResponse) []string {
	if resp == nil || len(resp.Data) == 0 {
		return nil
	}

	urls := make([]string, 0, len(resp.Data))
	for _, data := range resp.Data {
		if data.URL != "" {
			urls = append(urls, data.URL)
		}
	}

	return urls
}

// GetFirstImageURL 获取第一张图像的URL（便捷方法）
func GetFirstImageURL(resp *ImageGenerationResponse) string {
	if resp == nil || len(resp.Data) == 0 {
		return ""
	}
	return resp.Data[0].URL
}

// GetImageBase64Data 获取图像的Base64数据列表（便捷方法）
func GetImageBase64Data(resp *ImageGenerationResponse) []string {
	if resp == nil || len(resp.Data) == 0 {
		return nil
	}

	data := make([]string, 0, len(resp.Data))
	for _, img := range resp.Data {
		if img.B64JSON != "" {
			data = append(data, img.B64JSON)
		}
	}

	return data
}

// intPtr 返回int指针（辅助函数）
func intPtr(i int) *int {
	return &i
}

// float64Ptr 返回float64指针（辅助函数）
func float64Ptr(f float64) *float64 {
	return &f
}