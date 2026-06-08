package article

import (
	"context"
	"time"

	"english-study/internal/errors"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListArticleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewListArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *ListArticleLogic {
	return &ListArticleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

// ListArticle 收录列表: 分页 + 关键词(标题中英 + 含词英文) + 标签多选; 批量算含词/标签避免 N+1.
func (l *ListArticleLogic) ListArticle(req *types.ListArticleReq) (resp *types.ListArticleResp, err error) {
	m := l.svcCtx.Model
	uid := l.ui.ID

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	items, total, qerr := m.ListArticles(l.ctx, uid, req.Keyword, req.TagIDs, offset, limit)
	if qerr != nil {
		return nil, errors.ErrorDatabaseQueryError("查询文章列表失败").WithCause(qerr)
	}

	ids := make([]uint, 0, len(items))
	for _, a := range items {
		ids = append(ids, a.ID)
	}
	wordMap, werr := m.ArticleWordsByArticleIDs(l.ctx, uid, ids)
	if werr != nil {
		return nil, errors.ErrorDatabaseQueryError("查询文章含词失败").WithCause(werr)
	}
	tagMap, terr := m.ArticleTagsByArticleIDs(l.ctx, uid, ids)
	if terr != nil {
		return nil, errors.ErrorDatabaseQueryError("计算标签失败").WithCause(terr)
	}

	data := make([]*types.ArticleListItem, 0, len(items))
	for _, a := range items {
		words := wordMap[a.ID]
		if words == nil {
			words = []string{}
		}
		data = append(data, &types.ArticleListItem{
			ID:        a.ID,
			TitleEn:   a.TitleEn,
			TitleZh:   a.TitleZh,
			Tags:      toArticleTags(tagMap[a.ID]),
			Words:     words,
			CreatedAt: a.CreatedAt.Format(time.RFC3339),
		})
	}

	resp = &types.ListArticleResp{Data: data}
	resp.TotalCount = total
	resp.Offset = offset
	resp.Limit = limit
	return resp, nil
}
