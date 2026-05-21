package dictionary

import (
	"context"
	"fmt"
	"strings"

	"english-study/internal/errors"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchStardictLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewSearchStardictLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *SearchStardictLogic {
	return &SearchStardictLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *SearchStardictLogic) SearchStardict(req *types.SearchStardictReq) (*types.SearchStardictResp, error) {
	keyword := strings.TrimSpace(req.Keyword)
	if keyword == "" {
		return nil, errors.ErrorRequestParamError("keyword不能为空")
	}
	limit := req.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	// 转义 LIKE 元字符, 防 % / _ 通配符泄漏导致返回过多无关结果
	esc := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(keyword)

	sdg := l.svcCtx.Model.Gen.StarDict
	// 用 ILike 走 pg_trgm GIN 索引 (migration 006), 不全表扫
	results, err := sdg.WithContext(l.ctx).
		Where(sdg.Translation.Like("%" + esc + "%")).
		Order(sdg.Frq.Desc()).
		Limit(limit).
		Find()
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("搜索词库失败").WithCause(err)
	}

	if len(results) == 0 {
		return &types.SearchStardictResp{Data: []types.StardictItem{}}, nil
	}

	// 分类：含空格为短语，否则为单词
	var wordList, phraseList []string
	for _, r := range results {
		if strings.Contains(r.Sw, " ") {
			phraseList = append(phraseList, r.Sw)
		} else {
			wordList = append(wordList, r.Sw)
		}
	}

	// 批量查询已添加状态
	addedWords := make(map[string]bool)
	addedPhrases := make(map[string]bool)

	if len(wordList) > 0 {
		wordTable := fmt.Sprintf("word_user_%d", l.ui.ID)
		var existingWords []string
		if err := l.svcCtx.Model.DB.Table(wordTable).
			Where("word IN ?", wordList).
			Pluck("word", &existingWords).Error; err != nil {
			logx.Errorf("查询已添加单词失败: %v", err)
		}
		for _, w := range existingWords {
			addedWords[w] = true
		}
	}

	if len(phraseList) > 0 {
		phraseTable := fmt.Sprintf("word_phrase_user_%d", l.ui.ID)
		var existingPhrases []string
		if err := l.svcCtx.Model.DB.Table(phraseTable).
			Where("phrase IN ?", phraseList).
			Pluck("phrase", &existingPhrases).Error; err != nil {
			logx.Errorf("查询已添加短语失败: %v", err)
		}
		for _, p := range existingPhrases {
			addedPhrases[p] = true
		}
	}

	items := make([]types.StardictItem, 0, len(results))
	for _, r := range results {
		wordType := 1
		if strings.Contains(r.Sw, " ") {
			wordType = 2
		}
		items = append(items, types.StardictItem{
			Sw:          r.Sw,
			Phonetic:    r.Phonetic,
			Translation: r.Translation,
			WordType:    wordType,
			IsAdded:     addedWords[r.Sw] || addedPhrases[r.Sw],
		})
	}

	return &types.SearchStardictResp{Data: items}, nil
}
