// Package wordselect 抽取自 practise 包的选词/转卡公共逻辑, 供 practise 与 article 等多个业务包复用.
// 这些函数原本是 practise 包私有, 为避免跨包重复实现而上提并导出; 行为与原实现保持一致.
package wordselect

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
	"strings"
)

// WordStatusToWordCard 把一条 word_status 记录补齐为完整的 WordCard (按类型查单词或短语).
func WordStatusToWordCard(ctx context.Context, m *model.Model, w *bean.WordStatus) (*types.WordCard, error) {
	switch w.WordType {
	case types.WordTypeWord:
		bw, err := m.GetWordWithPosById(ctx, w.WordID, &w.UserID)
		if err != nil {
			return nil, fmt.Errorf("查询单词 %d 失败: %w", w.WordID, err)
		}
		return BeanWordToWordCard(bw), nil
	case types.WordTypePhrase:
		bwp, err := m.GetWordPhraseById(ctx, w.WordID, &w.UserID)
		if err != nil {
			return nil, fmt.Errorf("查询短语 %d 失败: %w", w.WordID, err)
		}
		return BeanPhraseToWordCard(bwp), nil
	default:
		return nil, fmt.Errorf("未知的单词类型 %d", w.WordType)
	}
}

// BeanWordToWordCard 单词 bean (含词性) -> WordCard.
func BeanWordToWordCard(w *bean.Word) *types.WordCard {
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

// BeanPhraseToWordCard 短语 bean -> WordCard.
func BeanPhraseToWordCard(w *bean.WordPhrase) *types.WordCard {
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

// ApplyTagFilter 在 word_statuses 查询链上把"必须打了任一指定标签"作为过滤条件 (OR 语义).
// 实现: 先单独查 word_tags 拿匹配的 (word_id, word_type) 对应的 word_statuses.id, 再用 In 限制主查询.
// tagIDs 为空时不动 (不筛选). 索引: word_tags(user_id, tag_id, word_type, word_id) 见迁移 009.
//
// 不直接拼 EXISTS 子查询, 是因为 GORM Gen 的 typed chain 不便直接挂 raw SQL; 这里两段查询
// 在单用户量级 (一般 < 10k 词条) 下成本可忽略.
func ApplyTagFilter(ctx context.Context, m *model.Model, find dto.IWordStatusDo, userID uint, tagIDs []uint) (dto.IWordStatusDo, error) {
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

	// 2. 反查 word_statuses 拿匹配的 id 列表.
	// 必须按"精确 (word_id, word_type) 对"匹配, 不能用 word_id IN (...) AND word_type IN (...):
	// 单词与短语的数字 id 各自独立, 二者可能撞号(如单词 235=blister 与短语 235=Alzheimer's),
	// 笛卡尔式 IN/IN 会把"撞号但没打该标签"的词误判进来.
	conds := make([]string, 0, len(rows))
	args := make([]interface{}, 0, len(rows)*2+1)
	args = append(args, userID)
	for _, r := range rows {
		conds = append(conds, "(word_id = ? AND word_type = ?)")
		args = append(args, r.WordID, r.WordType)
	}
	var ids []uint
	if err := m.DB.WithContext(ctx).
		Table("word_statuses").
		Where("user_id = ? AND ("+strings.Join(conds, " OR ")+")", args...).
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

// GetRandomWordStatus 从查询链里安全随机取 count 条: 取 count*3 条 + 随机偏移做样本, 再洗牌选 count.
func GetRandomWordStatus(find dto.IWordStatusDo, count int) (wss []*bean.WordStatus, err error) {
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
	return SafeRandomSelect(samples, count), nil
}

// SafeRandomSelect 用洗牌算法从 samples 里安全选 count 条 (避免越界).
func SafeRandomSelect(samples []*bean.WordStatus, count int) []*bean.WordStatus {
	if len(samples) <= count {
		return samples
	}

	shuffled := make([]*bean.WordStatus, len(samples))
	copy(shuffled, samples)

	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[:count]
}
