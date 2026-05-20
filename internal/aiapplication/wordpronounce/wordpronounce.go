package wordpronounce

import (
	"context"
)

// WordPronounce 是生成单词发音的接口
type WordPronounce interface {
	// GeneratePronounce 生成单词的发音
	GeneratePronounce(ctx context.Context, word string, opts ...OptionFunc) ([]byte, error)
	// GeneratePronouncePhonetic 生成单词的音标
	GeneratePronouncePhonetic(ctx context.Context, word string, opts ...OptionFunc) (string, error)
}
