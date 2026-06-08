package article

import (
	"context"
	"encoding/json"
	"strings"

	"english-study/internal/errors"
	"english-study/internal/model/bean"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveArticleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewSaveArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *SaveArticleLogic {
	return &SaveArticleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

// SaveArticle 收录: 事务写 article(标题+body JSON) + article_words(每个有 id 的词一行, 去重).
func (l *SaveArticleLogic) SaveArticle(req *types.SaveArticleReq) (resp *types.SaveArticleResp, err error) {
	if strings.TrimSpace(req.TitleEn) == "" || len(req.Sentences) == 0 {
		return nil, errors.ErrorRequestParamError("文章内容不完整")
	}

	body := storedBody{
		Sentences: make([]storedSentence, 0, len(req.Sentences)),
		UsedWords: make([]storedUsedWord, 0, len(req.UsedWords)),
	}
	for _, s := range req.Sentences {
		body.Sentences = append(body.Sentences, storedSentence{En: s.En, Zh: s.Zh})
	}
	for _, u := range req.UsedWords {
		body.UsedWords = append(body.UsedWords, storedUsedWord{Word: u.Word, Type: u.WordType, Surfaces: u.Surfaces})
	}
	raw, merr := json.Marshal(body)
	if merr != nil {
		return nil, errors.ErrorRequestParamError("序列化文章失败").WithCause(merr)
	}

	art := &bean.Article{
		UserID:  l.ui.ID,
		TitleEn: req.TitleEn,
		TitleZh: req.TitleZh,
		Body:    string(raw),
	}

	// article_words: 仅对有 word_id 的词建关联行, 按 (word_id, word_type) 去重
	seen := make(map[string]struct{}, len(req.UsedWords))
	words := make([]bean.ArticleWord, 0, len(req.UsedWords))
	for _, u := range req.UsedWords {
		if u.WordID == 0 {
			continue
		}
		k := pairKey(u.WordID, u.WordType)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		words = append(words, bean.ArticleWord{WordID: u.WordID, WordType: u.WordType, WordText: u.Word})
	}

	id, ierr := l.svcCtx.Model.CreateArticleWithWords(l.ctx, art, words)
	if ierr != nil {
		return nil, errors.ErrorDatabaseInsertError("收录文章失败").WithCause(ierr)
	}
	return &types.SaveArticleResp{ID: id}, nil
}
