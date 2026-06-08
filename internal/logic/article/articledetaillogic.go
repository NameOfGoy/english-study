package article

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"time"

	"english-study/internal/errors"
	"english-study/internal/model"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type ArticleDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewArticleDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *ArticleDetailLogic {
	return &ArticleDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

// ArticleDetail 详情: 解析 body JSON, 读时补齐 used_words 简要信息(词被删则留空), 实时算标签并集. 渲染体与即时文章一致.
func (l *ArticleDetailLogic) ArticleDetail(req *types.ArticleDetailReq) (resp *types.ArticleDetailResp, err error) {
	m := l.svcCtx.Model
	uid := l.ui.ID

	art, err := m.GetArticleByID(l.ctx, req.ID, uid)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrorWordNotExistError("文章不存在")
		}
		return nil, errors.ErrorDatabaseQueryError("查询文章失败").WithCause(err)
	}

	var body storedBody
	if art.Body != "" {
		_ = json.Unmarshal([]byte(art.Body), &body)
	}

	aws, awerr := m.GetArticleWords(l.ctx, req.ID, uid)
	if awerr != nil {
		return nil, errors.ErrorDatabaseQueryError("查询文章词条失败").WithCause(awerr)
	}
	idIndex := make(map[string]uint, len(aws))
	for _, w := range aws {
		idIndex[pairKeyText(w.WordText, w.WordType)] = w.WordID
	}
	// 只保留确实收录在 article_words 里的词, 过滤掉历史数据中 AI 误加, 未入 article_words 的额外词
	// (例如 peeler 衍生出的 police officer): 避免这些词被高亮.
	keptUsed := make([]storedUsedWord, 0, len(body.UsedWords))
	for _, u := range body.UsedWords {
		if _, ok := idIndex[pairKeyText(u.Word, u.Type)]; ok {
			keptUsed = append(keptUsed, u)
		}
	}
	lookup := func(word string, t int) (uint, model.WordBrief) {
		return idIndex[pairKeyText(word, t)], m.GetWordBrief(l.ctx, uid, word, t)
	}

	tagMap, terr := m.ArticleTagsByArticleIDs(l.ctx, uid, []uint{art.ID})
	if terr != nil {
		return nil, errors.ErrorDatabaseQueryError("计算标签失败").WithCause(terr)
	}

	return &types.ArticleDetailResp{
		Data: types.ArticleView{
			ID:        art.ID,
			TitleEn:   art.TitleEn,
			TitleZh:   art.TitleZh,
			Tags:      toArticleTags(tagMap[art.ID]),
			Sentences: storedSentencesToTypes(body.Sentences),
			UsedWords: buildUsedWordsFromStored(body.UsedWords, lookup),
			CreatedAt: art.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}
