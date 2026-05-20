package bigmodel

import (
	"context"
	"encoding/base64"
	"english-study/internal/utils"
	"fmt"
	"io"
	"net/http"
	"time"

	"english-study/internal/AI/view"
	"english-study/internal/thirdpart/bigmodel"
)

const (
	defaultModel = bigmodel.ModelCogView3Flash
)

// BigModelView 智谱清言图像生成实现
type BigModelView struct {
	sdk *bigmodel.SDK
}

// NewBigModelView 创建BigModelView实例
func NewBigModelView(apiKey string) *BigModelView {
	return &BigModelView{
		sdk: bigmodel.New(apiKey),
	}
}

// NewBigModelViewWithConfig 使用自定义配置创建BigModelView实例
func NewBigModelViewWithConfig(apiKey, baseURL string, timeout time.Duration) *BigModelView {
	return &BigModelView{
		sdk: bigmodel.NewWithConfig(apiKey, baseURL, timeout),
	}
}

// Generate 生成图像
func (b *BigModelView) Generate(ctx context.Context, description string, opts ...view.Options) ([]byte, error) {
	// 应用选项配置
	options := utils.GetOptionFromOptions[view.Option, view.Options](opts, func() view.Option {
		return view.Option{
			Model:          defaultModel,
			Size:           view.ImageSizeLarge,
			Quality:        view.ImageQualityStandard,
			Count:          1,
			ResponseFormat: view.ResponseFormatURL,
		}
	})

	// 验证参数
	if err := b.validateOptions(&options); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	// 构建请求
	req := b.buildRequest(description, &options)

	// 调用SDK生成图像
	resp, err := b.sdk.Image().CreateImage(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate image: %w", err)
	}

	// 处理响应
	return b.handleResponse(resp, &options)
}

// validateOptions 验证选项参数
func (b *BigModelView) validateOptions(options *view.Option) error {
	if options.Model == "" {
		return fmt.Errorf("model is required")
	}

	if options.Count < 1 || options.Count > 10 {
		return fmt.Errorf("count must be between 1 and 10, got %d", options.Count)
	}

	// 验证图像尺寸格式
	if !b.isValidSize(string(options.Size)) {
		return fmt.Errorf("invalid image size: %s", options.Size)
	}

	return nil
}

// isValidSize 检查图像尺寸是否有效
func (b *BigModelView) isValidSize(size string) bool {
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

// buildRequest 构建SDK请求
func (b *BigModelView) buildRequest(description string, options *view.Option) bigmodel.ImageGenerationRequest {
	req := bigmodel.ImageGenerationRequest{
		Model:  options.Model,
		Prompt: description,
		Size:   string(options.Size),
		N:      &options.Count,
	}

	// 设置质量参数
	if options.Quality != "" {
		req.Quality = string(options.Quality)
	}

	return req
}

// handleResponse 处理SDK响应
func (b *BigModelView) handleResponse(resp *bigmodel.ImageGenerationResponse, options *view.Option) ([]byte, error) {
	if resp == nil || len(resp.Data) == 0 {
		return nil, fmt.Errorf("no image data in response")
	}

	// 获取第一张图像数据
	imageData := resp.Data[0]

	switch options.ResponseFormat {
	case view.ResponseFormatBase64:
		// 如果SDK直接返回Base64数据
		if imageData.B64JSON != "" {
			return base64.StdEncoding.DecodeString(imageData.B64JSON)
		}
		// 如果是URL，需要下载并转换为Base64
		if imageData.URL != "" {
			return b.downloadImageAsBytes(imageData.URL)
		}
		return nil, fmt.Errorf("no base64 or URL data available")
	case view.ResponseFormatURL:
		if imageData.URL != "" {
			return []byte(imageData.URL), nil
		}
		return nil, fmt.Errorf("no URL data available")
	default:
		// 默认返回URL对应的图像字节数据
		if imageData.URL != "" {
			return b.downloadImageAsBytes(imageData.URL)
		}
		// 如果有Base64数据，解码返回
		if imageData.B64JSON != "" {
			return base64.StdEncoding.DecodeString(imageData.B64JSON)
		}
		return nil, fmt.Errorf("no image data available")
	}
}

// downloadImageAsBytes 下载图像并返回字节数据
func (b *BigModelView) downloadImageAsBytes(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
