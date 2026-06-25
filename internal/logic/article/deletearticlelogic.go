package article

import (
	"context"

	"english-study/internal/errors"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteArticleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewDeleteArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *DeleteArticleLogic {
	return &DeleteArticleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

// DeleteArticle 删除一篇收录文章(连同 article_words). 仅能删本人文章, 越权/不存在均返回未找到.
func (l *DeleteArticleLogic) DeleteArticle(req *types.DeleteArticleReq) (resp *types.DeleteArticleResp, err error) {
	if req.ID == 0 {
		return nil, errors.ErrorRequestParamError("文章ID不能为空")
	}

	affected, derr := l.svcCtx.Model.DeleteArticle(l.ctx, req.ID, l.ui.ID)
	if derr != nil {
		return nil, errors.ErrorDatabaseDeleteError("删除文章失败").WithCause(derr)
	}
	if affected == 0 {
		return nil, errors.ErrorDatabaseQueryError("文章不存在或无权删除")
	}

	return &types.DeleteArticleResp{}, nil
}
