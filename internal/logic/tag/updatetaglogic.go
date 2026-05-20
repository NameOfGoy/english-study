package tag

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/model/bean"
	"english-study/internal/utils"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTagLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewUpdateTagLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *UpdateTagLogic {
	return &UpdateTagLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *UpdateTagLogic) UpdateTag(req *types.UpdateTagReq) (resp *types.UpdateTagResp, err error) {

	tg := l.svcCtx.Model.Gen.Tag
	// 更新标签
	_, err = tg.WithContext(l.ctx).Where(tg.ID.Eq(req.ID), tg.UserID.Eq(l.ui.ID)).Updates(&bean.Tag{
		ID:    req.ID,
		Tag:   req.Name,
		Style: req.Style,
	})
	if err != nil {
		return nil, errors.ErrorDatabaseUpdateError("更新标签失败").WithCause(err)
	}

	return
}
