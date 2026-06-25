package article

import (
	"context"

	"english-study/internal/errors"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

// maxBatchDelete 单次批量删除上限, 防止超大 IN 列表
const maxBatchDelete = 200

type BatchDeleteArticleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewBatchDeleteArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *BatchDeleteArticleLogic {
	return &BatchDeleteArticleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

// BatchDeleteArticle 批量删除收录文章(连同 article_words). 仅删本人文章, 去重后按 user_id 隔离.
func (l *BatchDeleteArticleLogic) BatchDeleteArticle(req *types.BatchDeleteArticleReq) (resp *types.BatchDeleteArticleResp, err error) {
	// 去重 + 过滤 0
	seen := make(map[uint]struct{}, len(req.IDs))
	ids := make([]uint, 0, len(req.IDs))
	for _, id := range req.IDs {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, errors.ErrorRequestParamError("未选择要删除的文章")
	}
	if len(ids) > maxBatchDelete {
		return nil, errors.ErrorRequestParamError("一次最多删除 200 篇文章")
	}

	affected, derr := l.svcCtx.Model.DeleteArticles(l.ctx, ids, l.ui.ID)
	if derr != nil {
		return nil, errors.ErrorDatabaseDeleteError("批量删除文章失败").WithCause(derr)
	}

	return &types.BatchDeleteArticleResp{Deleted: int(affected)}, nil
}
