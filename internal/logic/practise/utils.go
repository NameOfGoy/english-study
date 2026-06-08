package practise

import (
	"context"
	"english-study/internal/logic/wordselect"
	"english-study/internal/model"
	"english-study/internal/model/bean"
	"english-study/internal/model/dto"
	"english-study/internal/types"
	"time"
)

// 以下为 internal/logic/wordselect 包的同名薄包装, 保持 practise 内既有调用点不变 (零行为变化).
// 选词/转卡的真正实现已上提到 wordselect 供 practise 与 article 等多包复用.

func wordStatusToWordCard(ctx context.Context, m *model.Model, w *bean.WordStatus) (*types.WordCard, error) {
	return wordselect.WordStatusToWordCard(ctx, m, w)
}

func beanWordToWordCard(w *bean.Word) *types.WordCard {
	return wordselect.BeanWordToWordCard(w)
}

func beanPhraseToWordCard(w *bean.WordPhrase) *types.WordCard {
	return wordselect.BeanPhraseToWordCard(w)
}

func applyTagFilter(ctx context.Context, m *model.Model, find dto.IWordStatusDo, userID uint, tagIDs []uint) (dto.IWordStatusDo, error) {
	return wordselect.ApplyTagFilter(ctx, m, find, userID, tagIDs)
}

func getRandomWordStatus(find dto.IWordStatusDo, count int) ([]*bean.WordStatus, error) {
	return wordselect.GetRandomWordStatus(find, count)
}

func safeRandomSelect(samples []*bean.WordStatus, count int) []*bean.WordStatus {
	return wordselect.SafeRandomSelect(samples, count)
}

// getTodayZeroTime 仅 SRS 复习逻辑使用, 不上提, 保留在 practise.
func getTodayZeroTime() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}
