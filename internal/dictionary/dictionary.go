package dictionary

import (
	"context"
	"english-study/internal/types"
)

// 字典接口
type Dictionary interface {
	// 新增单词到主词典；triggerUserID 用于异步生成的例句回填到该用户表（0=不回填）
	AddWord(ctx context.Context, word string, triggerUserID uint) error
	// 获取单词信息
	GetWord(ctx context.Context, word string) (*types.Word, error)
	// 新增短语
	AddPhrase(ctx context.Context, phrase string) error
	// 获取短语信息
	GetPhrase(ctx context.Context, phrase string) (*types.WordPhrase, error)
	// 词语是否在添加中
	IsWordAdding(ctx context.Context, word string) bool
	// 是否是单词
	IsWord(ctx context.Context, word string) bool
}
