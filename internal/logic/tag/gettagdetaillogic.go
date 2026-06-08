package tag

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/utils"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTagDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewGetTagDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *GetTagDetailLogic {
	return &GetTagDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *GetTagDetailLogic) GetTagDetail(req *types.GetTagDetailReq) (resp *types.GetTagDetailResp, err error) {

	tg := l.svcCtx.Model.Gen.Tag
	// 查询标签详情: 可见性 = 自己的私有标签 + 系统标签
	tag, err := tg.WithContext(l.ctx).Where(tg.ID.Eq(req.ID)).
		Where(tg.WithContext(l.ctx).Where(tg.IsSystem.Is(true)).Or(tg.UserID.Eq(l.ui.ID))).
		First()
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询标签详情失败").WithCause(err)
	}

	return &types.GetTagDetailResp{
		Data: types.Tag{
			ID:       tag.ID,
			Name:     tag.Tag,
			Style:    tag.Style,
			IsSystem: tag.IsSystem,
		},
	}, nil
}
