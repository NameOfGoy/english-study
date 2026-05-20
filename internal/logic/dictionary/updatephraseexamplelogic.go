package dictionary

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/model/bean"
	"english-study/internal/utils"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdatePhraseExampleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewUpdatePhraseExampleLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *UpdatePhraseExampleLogic {
	return &UpdatePhraseExampleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *UpdatePhraseExampleLogic) UpdatePhraseExample(req *types.UpdatePhraseExampleReq) (resp *types.UpdatePhraseExampleResp, err error) {
	err = l.svcCtx.Model.DB.Table((&bean.WordPhrase{}).UserTableName(&l.ui.ID)).
		Where("id = ?", req.ID).WithContext(l.ctx).
		Updates(map[string]interface{}{
			"example": (&types.WordPhrase{Example: req.Examples}).ExampleString(),
		}).Error
	if err != nil {
		return nil, errors.ErrorDatabaseUpdateError("更新例句失败").WithCause(err)
	}
	return &types.UpdatePhraseExampleResp{}, nil
}
