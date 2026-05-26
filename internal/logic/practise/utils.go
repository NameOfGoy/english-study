package practise

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/model"
	"english-study/internal/model/bean"
	"english-study/internal/model/dto"
	"english-study/internal/types"
	"english-study/internal/utils"
	"fmt"
	"math/rand"
	"time"
)

func wordStatusToWordCard(ctx context.Context, m *model.Model, w *bean.WordStatus) (*types.WordCard, error) {
	switch w.WordType {
	case types.WordTypeWord:
		bw, err := m.GetWordWithPosById(ctx, w.WordID, &w.UserID)
		if err != nil {
			return nil, fmt.Errorf("查询单词 %d 失败: %w", w.WordID, err)
		}
		return beanWordToWordCard(bw), nil
	case types.WordTypePhrase:
		bwp, err := m.GetWordPhraseById(ctx, w.WordID, &w.UserID)
		if err != nil {
			return nil, fmt.Errorf("查询短语 %d 失败: %w", w.WordID, err)
		}
		return beanPhraseToWordCard(bwp), nil
	default:
		return nil, fmt.Errorf("未知的单词类型 %d", w.WordType)
	}
}

func beanWordToWordCard(w *bean.Word) *types.WordCard {
	var translation string
	var example []string
	var picture []string
	var picturePosIds []uint
	var transItems []types.TranslationItem
	for _, pos := range w.Pos {
		if pos.Translation != "" {
			translation += fmt.Sprintf("%s %s", types.ToPosSw(pos.Pos), pos.Translation) + "\n"
		}
		if pos.Example != "" {
			example = append(example, pos.Example)
		}
		if pos.Picture != "" {
			picture = append(picture, utils.ToOssUri(types.OssBucket, pos.Picture))
			picturePosIds = append(picturePosIds, pos.ID)
		}
		transItems = append(transItems, types.TranslationItem{
			ID:          pos.ID,
			Pos:         pos.Pos,
			PosLabel:    types.ToPosSw(pos.Pos),
			Translation: pos.Translation,
		})
	}
	return &types.WordCard{
		ID:               w.ID,
		Word:             w.Word,
		WordType:         types.WordTypeWord,
		UKPhonetic:       w.BritishPronunciation,
		UKAudio:          utils.ToOssUri(types.OssBucket, w.BritishPronunciationAudio),
		USPhonetic:       w.AmericanPronunciation,
		USAudio:          utils.ToOssUri(types.OssBucket, w.AmericanPronunciationAudio),
		Translation:      translation,
		Example:          example,
		Picture:          picture,
		PicturePosIds:    picturePosIds,
		TranslationItems: transItems,
	}
}

func beanPhraseToWordCard(w *bean.WordPhrase) *types.WordCard {
	return &types.WordCard{
		ID:            w.ID,
		Word:          w.Phrase,
		WordType:      types.WordTypePhrase,
		UKAudio:       utils.ToOssUri(types.OssBucket, w.Pronunciation),
		Translation:   w.Translation,
		Example:       []string{w.Example},
		Picture:       []string{utils.ToOssUri(types.OssBucket, w.Picture)},
		PicturePosIds: []uint{w.ID},
		TranslationItems: []types.TranslationItem{
			{ID: w.ID, Pos: 0, PosLabel: "", Translation: w.Translation},
		},
	}
}

// applyTagFilter 在 word_statuses 查询链上把"必须打了任一指定标签"作为过滤条件.
// 实现: 先单独查 word_tags 拿匹配的 (word_id, word_type) 对应的 word_statuses.id, 再用 In 限制主查询.
// tagIDs 为空时不动 (不筛选). 索引: word_tags(user_id, tag_id, word_type, word_id) 见迁移 009.
//
// 不直接拼 EXISTS 子查询, 是因为 GORM Gen 的 typed chain 不便直接挂 raw SQL; 这里两段查询
// 在单用户量级 (一般 < 10k 词条) 下成本可忽略.
func applyTagFilter(ctx context.Context, m *model.Model, find dto.IWordStatusDo, userID uint, tagIDs []uint) (dto.IWordStatusDo, error) {
	if len(tagIDs) == 0 {
		return find, nil
	}
	// 1. 拿到当前用户 + 指定标签下涉及的 (word_id, word_type) 集合
	type wtKey struct {
		WordID   uint
		WordType int
	}
	var rows []wtKey
	if err := m.DB.WithContext(ctx).
		Table("word_tags").
		Where("user_id = ? AND tag_id IN ?", userID, tagIDs).
		Select("DISTINCT word_id, word_type").
		Find(&rows).Error; err != nil {
		return find, err
	}
	if len(rows) == 0 {
		// 没有任何匹配, 返回一个永远空的查询
		wsg := m.Gen.WordStatus
		return find.Where(wsg.ID.Eq(0)), nil
	}

	// 2. 反查 word_statuses 拿匹配的 id 列表
	wordIDs := make([]uint, 0, len(rows))
	wordTypes := make([]int, 0, len(rows))
	for _, r := range rows {
		wordIDs = append(wordIDs, r.WordID)
		wordTypes = append(wordTypes, r.WordType)
	}
	var ids []uint
	if err := m.DB.WithContext(ctx).
		Table("word_statuses").
		Where("user_id = ? AND word_id IN ? AND word_type IN ?", userID, wordIDs, wordTypes).
		Pluck("id", &ids).Error; err != nil {
		return find, err
	}
	if len(ids) == 0 {
		wsg := m.Gen.WordStatus
		return find.Where(wsg.ID.Eq(0)), nil
	}
	wsg := m.Gen.WordStatus
	return find.Where(wsg.ID.In(ids...)), nil
}

func getRandomWordStatus(find dto.IWordStatusDo, count int) (wss []*bean.WordStatus, err error) {

	// 随机, 取 req.Count * 3 条数据, 随机偏移量; 再从这些数据中随机取 req.Count 条

	// 先看看总数
	total, err := find.Count()
	if err != nil {
		return
	}
	if total <= int64(count) {
		return find.Find()
	}
	// 安全计算样本数
	sc := min(int(total), count*3)

	var offset int
	if int(total) > sc {
		offset = rand.Intn(int(total) - sc)
	}

	samples, err := find.Limit(sc).Offset(offset).Find()
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询学习中的单词失败").WithCause(err)
	}

	// 使用安全的随机选择方法
	return safeRandomSelect(samples, count), nil
}

// 安全的随机选择函数
func safeRandomSelect(samples []*bean.WordStatus, count int) []*bean.WordStatus {
	if len(samples) <= count {
		return samples
	}

	// 使用洗牌算法避免索引越界
	shuffled := make([]*bean.WordStatus, len(samples))
	copy(shuffled, samples)

	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[:count]
}

func getTodayZeroTime() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}
