package dictionary

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/model/bean"

	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateExampleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewUpdateExampleLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *UpdateExampleLogic {
	return &UpdateExampleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *UpdateExampleLogic) UpdateExample(req *types.UpdateExampleReq) (resp *types.UpdateExampleResp, err error) {
	err = l.svcCtx.Model.DB.Table((&bean.WordPos{}).UserTableName(&l.ui.ID)).
		Where("id = ?", req.WordPosId).WithContext(l.ctx).
		Updates(map[string]interface{}{
			"example": (&types.WordPos{Example: req.Examples}).ExampleString(),
		}).Error
	if err != nil {
		return nil, errors.ErrorDatabaseUpdateError("更新例句失败").WithCause(err)
	}

	return &types.UpdateExampleResp{}, nil
}
