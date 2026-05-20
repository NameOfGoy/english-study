package wordexample

import (
	"context"
	"english-study/internal/types"
)

// WordExample 是生成英语单词例句的接口
type WordExample interface {
	// Generate 根据词语生成例句
	// word: 词语
	// opts: 可选参数
	Generate(ctx context.Context, word string, opts ...OptionFunc) ([]*types.Example, error)
}
