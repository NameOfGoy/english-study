package video

import "context"

// Video 视频生成接口
type Video interface {
	// Generate 生成视频
	Generate(ctx context.Context, description string, opts ...Option) ([]byte, error)
}
