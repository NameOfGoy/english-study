package dictionary

import (
	"context"
	"english-study/internal/utils"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExportWordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

// 字典模块/批量导出单词
func NewExportWordLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *ExportWordLogic {
	return &ExportWordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *ExportWordLogic) ExportWord(req *types.ExportWordReq) (resp *types.ExportWordResp, err error) {
	// todo: add your logic here and delete this line

	return
}
