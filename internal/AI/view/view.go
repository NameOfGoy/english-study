package view

import "context"

// View 图片生成接口
type View interface {
	// Generate 生成接口; 返回值是图片的字节数组
	Generate(ctx context.Context, description string, opts ...Options) ([]byte, error)
}
