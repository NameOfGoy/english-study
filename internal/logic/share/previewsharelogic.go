package share

import (
	"context"
	"fmt"

	"english-study/internal/errors"
	"english-study/internal/model/bean"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type PreviewShareLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewPreviewShareLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *PreviewShareLogic {
	return &PreviewShareLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *PreviewShareLogic) PreviewShare(req *types.PreviewShareReq) (*types.PreviewShareResp, error) {
	payload, err := DecodeToken(req.Token, l.svcCtx.Config.Auth.AccessSecret)
	if err != nil {
		return nil, errors.ErrorRequestParamError("分享码无效或已过期").WithCause(err)
	}
	sourceUserID := uint(payload.UserID)
	if sourceUserID == l.ui.ID {
		return nil, errors.ErrorRequestParamError("不能导入自己的分享码")
	}

	// 查 A 的用户名
	ug := l.svcCtx.Model.Gen.User
	srcUser, err := ug.WithContext(l.ctx).Where(ug.ID.Eq(sourceUserID)).Take()
	if err != nil {
		return nil, errors.ErrorRequestParamError("分享者不存在").WithCause(err)
	}

	resp := &types.PreviewShareResp{
		FromUsername: srcUser.Username,
		ExpiresAt:    int64(payload.ExpireTs),
		Items:        []types.ShareWordItem{},
		Tags:         []types.SharePreviewTag{},
	}

	wordIDs, phraseIDs, err := collectShareItemIDs(l.ctx, l.svcCtx, sourceUserID, payload)
	if err != nil {
		return nil, err
	}
	resp.WordCount = len(wordIDs)
	resp.PhraseCount = len(phraseIDs)

	// 准备 tag 映射 (wordType, itemID) -> []tagName
	tagsByItem, err := l.fetchTagsByItem(sourceUserID, wordIDs, phraseIDs)
	if err != nil {
		logx.Errorf("查询条目标签失败: %v", err)
		tagsByItem = map[string][]string{}
	}

	// 单词条目
	if len(wordIDs) > 0 {
		items, err := l.fetchWordItems(sourceUserID, wordIDs, tagsByItem)
		if err != nil {
			return nil, err
		}
		resp.Items = append(resp.Items, items...)
	}

	// 短语条目
	if len(phraseIDs) > 0 {
		items, err := l.fetchPhraseItems(sourceUserID, phraseIDs, tagsByItem)
		if err != nil {
			return nil, err
		}
		resp.Items = append(resp.Items, items...)
	}

	// 整体涉及的标签清单（去重）
	if len(tagsByItem) > 0 {
		nameSet := map[string]bool{}
		for _, names := range tagsByItem {
			for _, n := range names {
				nameSet[n] = true
			}
		}
		tagNames := make([]string, 0, len(nameSet))
		for n := range nameSet {
			tagNames = append(tagNames, n)
		}
		if len(tagNames) > 0 {
			var tags []bean.Tag
			l.svcCtx.Model.DB.WithContext(l.ctx).
				Where("user_id = ? AND tag IN ?", sourceUserID, tagNames).
				Find(&tags)
			for _, t := range tags {
				resp.Tags = append(resp.Tags, types.SharePreviewTag{
					Name:  t.Tag,
					Style: t.Style,
				})
			}
		}
	}

	return resp, nil
}

// fetchTagsByItem 一次性查出每个 word/phrase 对应的标签名
// 返回 key 格式 "1_{id}" / "2_{id}" -> [tagNames]
func (l *PreviewShareLogic) fetchTagsByItem(sourceUserID uint, wordIDs, phraseIDs []uint) (map[string][]string, error) {
	result := map[string][]string{}

	// 取所有相关 word_tags 关系
	var relations []bean.WordTag
	q := l.svcCtx.Model.DB.WithContext(l.ctx).Where("user_id = ?", sourceUserID)
	if len(wordIDs) > 0 && len(phraseIDs) > 0 {
		q = q.Where("(word_type = 1 AND word_id IN ?) OR (word_type = 2 AND word_id IN ?)", wordIDs, phraseIDs)
	} else if len(wordIDs) > 0 {
		q = q.Where("word_type = 1 AND word_id IN ?", wordIDs)
	} else if len(phraseIDs) > 0 {
		q = q.Where("word_type = 2 AND word_id IN ?", phraseIDs)
	} else {
		return result, nil
	}
	if err := q.Find(&relations).Error; err != nil {
		return result, err
	}
	if len(relations) == 0 {
		return result, nil
	}

	// 取 tag 详情
	tagIDSet := map[uint]bool{}
	for _, r := range relations {
		tagIDSet[r.TagID] = true
	}
	tagIDs := make([]uint, 0, len(tagIDSet))
	for id := range tagIDSet {
		tagIDs = append(tagIDs, id)
	}
	var tags []bean.Tag
	if err := l.svcCtx.Model.DB.WithContext(l.ctx).
		Where("id IN ? AND user_id = ?", tagIDs, sourceUserID).
		Find(&tags).Error; err != nil {
		return result, err
	}
	tagNameByID := map[uint]string{}
	for _, t := range tags {
		tagNameByID[t.ID] = t.Tag
	}

	for _, r := range relations {
		name, ok := tagNameByID[r.TagID]
		if !ok {
			continue
		}
		key := fmt.Sprintf("%d_%d", r.WordType, r.WordID)
		result[key] = append(result[key], name)
	}
	return result, nil
}

// fetchWordItems 把 A 的 word_user_{A} 行转成 ShareWordItem，附带第一个词性的中文翻译
func (l *PreviewShareLogic) fetchWordItems(sourceUserID uint, wordIDs []uint, tagsByItem map[string][]string) ([]types.ShareWordItem, error) {
	wordTable := fmt.Sprintf("word_user_%d", sourceUserID)
	posTable := fmt.Sprintf("word_pos_user_%d", sourceUserID)

	type wordRow struct {
		ID   uint
		Word string
	}
	var rows []wordRow
	if err := l.svcCtx.Model.DB.WithContext(l.ctx).
		Table(wordTable).
		Select("id, word").
		Where("id IN ?", wordIDs).
		Find(&rows).Error; err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询A单词失败").WithCause(err)
	}

	// 取每个 word 的第一个 pos 翻译
	type posRow struct {
		WordID      uint
		Translation string
	}
	var posRows []posRow
	if err := l.svcCtx.Model.DB.WithContext(l.ctx).
		Table(posTable).
		Select("word_id, translation").
		Where("word_id IN ?", wordIDs).
		Find(&posRows).Error; err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询A词性翻译失败").WithCause(err)
	}
	transByWord := map[uint]string{}
	for _, p := range posRows {
		if _, ok := transByWord[p.WordID]; !ok {
			transByWord[p.WordID] = p.Translation
		}
	}

	items := make([]types.ShareWordItem, 0, len(rows))
	for _, r := range rows {
		key := fmt.Sprintf("1_%d", r.ID)
		items = append(items, types.ShareWordItem{
			ID:          r.ID,
			WordType:    1,
			Text:        r.Word,
			Translation: transByWord[r.ID],
			TagNames:    tagsByItem[key],
		})
	}
	return items, nil
}

// fetchPhraseItems 同上，针对短语
func (l *PreviewShareLogic) fetchPhraseItems(sourceUserID uint, phraseIDs []uint, tagsByItem map[string][]string) ([]types.ShareWordItem, error) {
	phraseTable := fmt.Sprintf("word_phrase_user_%d", sourceUserID)
	type phraseRow struct {
		ID          uint
		Phrase      string
		Translation string
	}
	var rows []phraseRow
	if err := l.svcCtx.Model.DB.WithContext(l.ctx).
		Table(phraseTable).
		Select("id, phrase, translation").
		Where("id IN ?", phraseIDs).
		Find(&rows).Error; err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询A短语失败").WithCause(err)
	}

	items := make([]types.ShareWordItem, 0, len(rows))
	for _, r := range rows {
		key := fmt.Sprintf("2_%d", r.ID)
		items = append(items, types.ShareWordItem{
			ID:          r.ID,
			WordType:    2,
			Text:        r.Phrase,
			Translation: r.Translation,
			TagNames:    tagsByItem[key],
		})
	}
	return items, nil
}

// collectShareItemIDs 根据 payload 收集 A 用户要分享的单词ID和短语ID列表
func collectShareItemIDs(ctx context.Context, svcCtx *svc.ServiceContext, sourceUserID uint, payload *SharePayload) (wordIDs []uint, phraseIDs []uint, err error) {
	wordIDs = []uint{}
	phraseIDs = []uint{}

	wantWord := payload.WordType == WordTypeAll || payload.WordType == WordTypeWord
	wantPhrase := payload.WordType == WordTypeAll || payload.WordType == WordTypePhrase

	if payload.ShareType == ShareTypeAll {
		if wantWord {
			wordTable := fmt.Sprintf("word_user_%d", sourceUserID)
			if e := svcCtx.Model.DB.WithContext(ctx).Table(wordTable).Pluck("id", &wordIDs).Error; e != nil {
				return nil, nil, errors.ErrorDatabaseQueryError("查询A用户单词列表失败").WithCause(e)
			}
		}
		if wantPhrase {
			phraseTable := fmt.Sprintf("word_phrase_user_%d", sourceUserID)
			if e := svcCtx.Model.DB.WithContext(ctx).Table(phraseTable).Pluck("id", &phraseIDs).Error; e != nil {
				return nil, nil, errors.ErrorDatabaseQueryError("查询A用户短语列表失败").WithCause(e)
			}
		}
		return
	}

	// ShareTypeByTag: 从 word_tags 过滤
	tagIDs := make([]uint, len(payload.TagIDs))
	for i, id := range payload.TagIDs {
		tagIDs[i] = uint(id)
	}
	if wantWord {
		if e := svcCtx.Model.DB.WithContext(ctx).
			Table("word_tags").
			Distinct("word_id").
			Where("user_id = ? AND word_type = 1 AND tag_id IN ?", sourceUserID, tagIDs).
			Pluck("word_id", &wordIDs).Error; e != nil {
			return nil, nil, errors.ErrorDatabaseQueryError("查询A用户按标签单词失败").WithCause(e)
		}
	}
	if wantPhrase {
		if e := svcCtx.Model.DB.WithContext(ctx).
			Table("word_tags").
			Distinct("word_id").
			Where("user_id = ? AND word_type = 2 AND tag_id IN ?", sourceUserID, tagIDs).
			Pluck("word_id", &phraseIDs).Error; e != nil {
			return nil, nil, errors.ErrorDatabaseQueryError("查询A用户按标签短语失败").WithCause(e)
		}
	}
	return
}
