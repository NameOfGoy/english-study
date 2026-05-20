package dictionary

import (
	"context"
	"english-study/internal/utils"

	"english-study/internal/svc"
	"english-study/internal/types"

	"english-study/internal/errors"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteWordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

// 字典模块/单词删除
func NewDeleteWordLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *DeleteWordLogic {
	return &DeleteWordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *DeleteWordLogic) DeleteWord(req *types.DeleteWordReq) (resp *types.DeleteWordResp, err error) {
	return nil, errors.ErrorNotSupportError("删除单词暂不支持")
}
