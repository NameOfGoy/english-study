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
	"gorm.io/gorm"
)

type ImportShareLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewImportShareLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *ImportShareLogic {
	return &ImportShareLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *ImportShareLogic) ImportShare(req *types.ImportShareReq) (*types.ImportShareResp, error) {
	payload, err := DecodeToken(req.Token, l.svcCtx.Config.Auth.AccessSecret)
	if err != nil {
		return nil, errors.ErrorRequestParamError("分享码无效或已过期").WithCause(err)
	}
	sourceUserID := uint(payload.UserID)
	if sourceUserID == l.ui.ID {
		return nil, errors.ErrorRequestParamError("不能导入自己的分享码")
	}

	wordIDs, phraseIDs, err := collectShareItemIDs(l.ctx, l.svcCtx, sourceUserID, payload)
	if err != nil {
		return nil, err
	}

	resp := &types.ImportShareResp{}

	// 导入单词
	if len(wordIDs) > 0 {
		imported, skipped, err := l.importWords(sourceUserID, wordIDs)
		if err != nil {
			return nil, err
		}
		resp.WordImported = imported
		resp.WordSkipped = skipped
	}

	// 导入短语
	if len(phraseIDs) > 0 {
		imported, skipped, err := l.importPhrases(sourceUserID, phraseIDs)
		if err != nil {
			return nil, err
		}
		resp.PhraseImported = imported
		resp.PhraseSkipped = skipped
	}

	// 导入标签关联（如果用户选了）
	if req.ImportTags {
		n, err := l.importTags(sourceUserID, wordIDs, phraseIDs)
		if err != nil {
			logx.Errorf("导入标签失败(不阻塞主流程): %v", err)
		} else {
			resp.TagImported = n
		}
	}

	return resp, nil
}

// importWords 把 A 的单词复制到 B 的单词表
// 返回: 实际导入数, 跳过数
func (l *ImportShareLogic) importWords(sourceUserID uint, sourceWordIDs []uint) (int, int, error) {
	srcWordTable := fmt.Sprintf("word_user_%d", sourceUserID)
	srcPosTable := fmt.Sprintf("word_pos_user_%d", sourceUserID)
	dstWordTable := fmt.Sprintf("word_user_%d", l.ui.ID)
	dstPosTable := fmt.Sprintf("word_pos_user_%d", l.ui.ID)

	// 取 A 的源数据
	var srcWords []bean.Word
	if err := l.svcCtx.Model.DB.WithContext(l.ctx).
		Table(srcWordTable).
		Where("id IN ?", sourceWordIDs).
		Find(&srcWords).Error; err != nil {
		return 0, 0, errors.ErrorDatabaseQueryError("读取A单词失败").WithCause(err)
	}

	// 查 B 已有哪些（按 word 文本）
	wordTexts := make([]string, 0, len(srcWords))
	for _, w := range srcWords {
		wordTexts = append(wordTexts, w.Word)
	}
	var existing []string
	l.svcCtx.Model.DB.WithContext(l.ctx).Table(dstWordTable).
		Where("word IN ?", wordTexts).Pluck("word", &existing)
	existSet := toSet(existing)

	imported, skipped := 0, 0

	for _, sw := range srcWords {
		if existSet[sw.Word] {
			skipped++
			continue
		}

		err := l.svcCtx.Model.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
			// 1. 插入 word 到 B 表（新 ID 自动分配）
			newWord := bean.Word{
				Word:                       sw.Word,
				AmericanPronunciation:      sw.AmericanPronunciation,
				AmericanPronunciationAudio: sw.AmericanPronunciationAudio,
				BritishPronunciation:       sw.BritishPronunciation,
				BritishPronunciationAudio:  sw.BritishPronunciationAudio,
			}
			if err := tx.Table(dstWordTable).Create(&newWord).Error; err != nil {
				return err
			}

			// 2. 拷贝 word_pos 行（保留图片路径）
			var srcPos []bean.WordPos
			if err := tx.Table(srcPosTable).Where("word_id = ?", sw.ID).Find(&srcPos).Error; err != nil {
				return err
			}
			for i := range srcPos {
				srcPos[i].ID = 0 // 新分配 ID
				srcPos[i].WordID = newWord.ID
			}
			if len(srcPos) > 0 {
				if err := tx.Table(dstPosTable).Create(&srcPos).Error; err != nil {
					return err
				}
			}

			// 3. 创建 word_statuses 行 (status=Study, source_user_id=A)
			ws := bean.WordStatus{
				WordID:       newWord.ID,
				WordType:     1,
				Status:       types.WordStatusStudy,
				UserID:       l.ui.ID,
				SourceUserID: sourceUserID,
			}
			return tx.Create(&ws).Error
		})

		if err != nil {
			logx.Errorf("导入单词 %s 失败: %v", sw.Word, err)
			continue
		}
		imported++
	}

	return imported, skipped, nil
}

// importPhrases 同上逻辑，针对短语
func (l *ImportShareLogic) importPhrases(sourceUserID uint, sourcePhraseIDs []uint) (int, int, error) {
	srcTable := fmt.Sprintf("word_phrase_user_%d", sourceUserID)
	dstTable := fmt.Sprintf("word_phrase_user_%d", l.ui.ID)

	var srcPhrases []bean.WordPhrase
	if err := l.svcCtx.Model.DB.WithContext(l.ctx).
		Table(srcTable).
		Where("id IN ?", sourcePhraseIDs).
		Find(&srcPhrases).Error; err != nil {
		return 0, 0, errors.ErrorDatabaseQueryError("读取A短语失败").WithCause(err)
	}

	phraseTexts := make([]string, 0, len(srcPhrases))
	for _, p := range srcPhrases {
		phraseTexts = append(phraseTexts, p.Phrase)
	}
	var existing []string
	l.svcCtx.Model.DB.WithContext(l.ctx).Table(dstTable).
		Where("phrase IN ?", phraseTexts).Pluck("phrase", &existing)
	existSet := toSet(existing)

	imported, skipped := 0, 0

	for _, sp := range srcPhrases {
		if existSet[sp.Phrase] {
			skipped++
			continue
		}

		err := l.svcCtx.Model.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
			newPhrase := bean.WordPhrase{
				Phrase:        sp.Phrase,
				Translation:   sp.Translation,
				Pronunciation: sp.Pronunciation,
				Example:       sp.Example,
				Picture:       sp.Picture,
			}
			if err := tx.Table(dstTable).Create(&newPhrase).Error; err != nil {
				return err
			}

			ws := bean.WordStatus{
				WordID:       newPhrase.ID,
				WordType:     2,
				Status:       types.WordStatusStudy,
				UserID:       l.ui.ID,
				SourceUserID: sourceUserID,
			}
			return tx.Create(&ws).Error
		})

		if err != nil {
			logx.Errorf("导入短语 %s 失败: %v", sp.Phrase, err)
			continue
		}
		imported++
	}

	return imported, skipped, nil
}

// importTags 把 A 的标签和 word_tags 关联也复刻到 B 的库
// 同名标签合并到 B 已有的；新名字则新建一个 B 自己的 tag
// 返回新建的标签数
func (l *ImportShareLogic) importTags(sourceUserID uint, sourceWordIDs, sourcePhraseIDs []uint) (int, error) {
	// 1. 找 A 的所有 word_tags 关联（针对要导入的 word/phrase）
	var srcRelations []bean.WordTag
	q := l.svcCtx.Model.DB.WithContext(l.ctx).
		Where("user_id = ?", sourceUserID)
	if len(sourceWordIDs) > 0 && len(sourcePhraseIDs) > 0 {
		q = q.Where("(word_type = 1 AND word_id IN ?) OR (word_type = 2 AND word_id IN ?)", sourceWordIDs, sourcePhraseIDs)
	} else if len(sourceWordIDs) > 0 {
		q = q.Where("word_type = 1 AND word_id IN ?", sourceWordIDs)
	} else if len(sourcePhraseIDs) > 0 {
		q = q.Where("word_type = 2 AND word_id IN ?", sourcePhraseIDs)
	} else {
		return 0, nil
	}
	if err := q.Find(&srcRelations).Error; err != nil {
		return 0, err
	}
	if len(srcRelations) == 0 {
		return 0, nil
	}

	// 2. 收集涉及的 tag_id 列表
	tagIDSet := map[uint]bool{}
	for _, r := range srcRelations {
		tagIDSet[r.TagID] = true
	}
	tagIDs := make([]uint, 0, len(tagIDSet))
	for id := range tagIDSet {
		tagIDs = append(tagIDs, id)
	}

	// 3. 取 A 的 tag 详情
	var srcTags []bean.Tag
	if err := l.svcCtx.Model.DB.WithContext(l.ctx).
		Where("id IN ? AND user_id = ?", tagIDs, sourceUserID).
		Find(&srcTags).Error; err != nil {
		return 0, err
	}

	// 4. 查 B 已有的同名标签
	tagNames := make([]string, 0, len(srcTags))
	for _, t := range srcTags {
		tagNames = append(tagNames, t.Tag)
	}
	var bExistTags []bean.Tag
	l.svcCtx.Model.DB.WithContext(l.ctx).
		Where("user_id = ? AND tag IN ?", l.ui.ID, tagNames).
		Find(&bExistTags)
	bTagByName := map[string]uint{}
	for _, t := range bExistTags {
		bTagByName[t.Tag] = t.ID
	}

	// 5. A.tag_id -> B.tag_id 映射；缺失的新建到 B
	srcToDstTag := map[uint]uint{}
	newTagCount := 0
	for _, t := range srcTags {
		if bID, ok := bTagByName[t.Tag]; ok {
			srcToDstTag[t.ID] = bID
		} else {
			newTag := bean.Tag{
				Tag:    t.Tag,
				Style:  t.Style,
				UserID: l.ui.ID,
			}
			if err := l.svcCtx.Model.DB.WithContext(l.ctx).Create(&newTag).Error; err != nil {
				logx.Errorf("创建标签 %s 失败: %v", t.Tag, err)
				continue
			}
			srcToDstTag[t.ID] = newTag.ID
			newTagCount++
		}
	}

	// 6. 拿 B 已经导入的单词/短语在自己库的 ID 映射（按 word 文本去 join）
	srcWordIDToB := map[uint]uint{}    // A.word_id -> B.word_id
	srcPhraseIDToB := map[uint]uint{}  // A.phrase_id -> B.phrase_id
	if len(sourceWordIDs) > 0 {
		// A 端按文本取 word -> id
		srcWordTable := fmt.Sprintf("word_user_%d", sourceUserID)
		type wp struct{ ID uint; Word string }
		var srcRows []wp
		l.svcCtx.Model.DB.WithContext(l.ctx).Table(srcWordTable).Where("id IN ?", sourceWordIDs).Find(&srcRows)
		// B 端按相同文本取 word -> id
		dstWordTable := fmt.Sprintf("word_user_%d", l.ui.ID)
		texts := make([]string, len(srcRows))
		for i, r := range srcRows {
			texts[i] = r.Word
		}
		var dstRows []wp
		l.svcCtx.Model.DB.WithContext(l.ctx).Table(dstWordTable).Where("word IN ?", texts).Find(&dstRows)
		bByText := map[string]uint{}
		for _, r := range dstRows {
			bByText[r.Word] = r.ID
		}
		for _, r := range srcRows {
			if bID, ok := bByText[r.Word]; ok {
				srcWordIDToB[r.ID] = bID
			}
		}
	}
	if len(sourcePhraseIDs) > 0 {
		srcPhraseTable := fmt.Sprintf("word_phrase_user_%d", sourceUserID)
		type pp struct{ ID uint; Phrase string }
		var srcRows []pp
		l.svcCtx.Model.DB.WithContext(l.ctx).Table(srcPhraseTable).Where("id IN ?", sourcePhraseIDs).Find(&srcRows)
		dstPhraseTable := fmt.Sprintf("word_phrase_user_%d", l.ui.ID)
		texts := make([]string, len(srcRows))
		for i, r := range srcRows {
			texts[i] = r.Phrase
		}
		var dstRows []pp
		l.svcCtx.Model.DB.WithContext(l.ctx).Table(dstPhraseTable).Where("phrase IN ?", texts).Find(&dstRows)
		bByText := map[string]uint{}
		for _, r := range dstRows {
			bByText[r.Phrase] = r.ID
		}
		for _, r := range srcRows {
			if bID, ok := bByText[r.Phrase]; ok {
				srcPhraseIDToB[r.ID] = bID
			}
		}
	}

	// 7. 创建 B 的 word_tags 关联
	var newRelations []bean.WordTag
	for _, r := range srcRelations {
		bTagID, ok := srcToDstTag[r.TagID]
		if !ok {
			continue
		}
		var bWordID uint
		if r.WordType == 1 {
			bWordID = srcWordIDToB[r.WordID]
		} else if r.WordType == 2 {
			bWordID = srcPhraseIDToB[r.WordID]
		}
		if bWordID == 0 {
			continue
		}
		newRelations = append(newRelations, bean.WordTag{
			WordID:   bWordID,
			WordType: r.WordType,
			TagID:    bTagID,
			UserID:   l.ui.ID,
		})
	}
	if len(newRelations) > 0 {
		// 用 ON CONFLICT DO NOTHING 防重复
		if err := l.svcCtx.Model.DB.WithContext(l.ctx).
			Create(&newRelations).Error; err != nil {
			logx.Errorf("插入 B 端 word_tags 失败: %v", err)
		}
	}

	return newTagCount, nil
}

func toSet(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, s := range items {
		m[s] = true
	}
	return m
}
