package impl

import (
	"context"
	"fmt"

	"english-study/internal/model/bean"
)

// aiFallbackToStarDict 当 stardict 没有该单词时，调 AI 生成 stardict-equivalent 数据
// 返回的 *bean.StarDict 是临时构造的，不会被插入 stardict 表
// 调用方应继续走正常 InsertWord 流程
func (d *DictionaryImpl) aiFallbackToStarDict(ctx context.Context, word string) (*bean.StarDict, error) {
	info, err := d.wordTranslationGenerator.GenerateWordInfo(ctx, word)
	if err != nil {
		return nil, fmt.Errorf("AI 生成单词信息失败: %w", err)
	}
	if !info.Valid {
		return nil, fmt.Errorf("[%s] 不是合法英语单词: %s", word, info.Reason)
	}

	return &bean.StarDict{
		Word:        word,
		Sw:          word,
		Phonetic:    info.Phonetic,
		Translation: info.Translation,
		Definition:  info.Definition,
		Exchange:    info.Exchange,
	}, nil
}
