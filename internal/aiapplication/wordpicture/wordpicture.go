package wordpicture

import (
	"context"
)

// Picture 是生成图片的接口
type Picture interface {
	// Generate 根据单词和词性生成图片
	// word: 单词
	// opts: 可选参数
	Generate(ctx context.Context, word string, opts ...OptionFunc) ([]byte, error)
}
