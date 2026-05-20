package view

import "fmt"

// ImageSize 图像尺寸类型
type ImageSize string

const (
	// 预设图像尺寸
	ImageSizeSmall     ImageSize = "256x256"   // 小尺寸
	ImageSizeMedium    ImageSize = "512x512"   // 中等尺寸
	ImageSizeLarge     ImageSize = "1024x1024" // 大尺寸
	ImageSizePortrait  ImageSize = "1024x1792" // 竖屏
	ImageSizeLandscape ImageSize = "1792x1024" // 横屏
)

// ImageQuality 图像质量类型
type ImageQuality string

const (
	ImageQualityStandard ImageQuality = "standard" // 标准质量
	ImageQualityHD       ImageQuality = "hd"       // 高清质量
)

// ResponseFormat 响应格式类型
type ResponseFormat string

const (
	ResponseFormatURL    ResponseFormat = "url"    // 返回图像URL
	ResponseFormatBase64 ResponseFormat = "base64" // 返回Base64编码
)

// Options 定义了View的选项
type Option struct {
	Model          string         // 模型名称
	Size           ImageSize      // 图像尺寸
	Quality        ImageQuality   // 图像质量
	Count          int            // 生成图像数量
	ResponseFormat ResponseFormat // 响应格式
}

// Option 是用于设置Options的函数类型
type Options func(*Option)

// WithModel 设置模型
func WithModel(model string) Options {
	return func(o *Option) {
		o.Model = model
	}
}

// WithSize 设置图像尺寸
func WithSize(size ImageSize) Options {
	return func(o *Option) {
		o.Size = size
	}
}

// WithCustomSize 设置自定义图像尺寸
func WithCustomSize(width, height int) Options {
	return func(o *Option) {
		o.Size = ImageSize(fmt.Sprintf("%dx%d", width, height))
	}
}

// WithQuality 设置图像质量
func WithQuality(quality ImageQuality) Options {
	return func(o *Option) {
		o.Quality = quality
	}
}

// WithCount 设置生成图像数量
func WithCount(count int) Options {
	return func(o *Option) {
		o.Count = count
	}
}

// WithResponseFormat 设置响应格式
func WithResponseFormat(format ResponseFormat) Options {
	return func(o *Option) {
		o.ResponseFormat = format
	}
}
