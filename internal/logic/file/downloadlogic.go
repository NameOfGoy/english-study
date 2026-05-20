package file

import (
	"context"
	"english-study/internal/utils"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DownloadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

// 文件模块/文件下载
func NewDownloadLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *DownloadLogic {
	return &DownloadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *DownloadLogic) Download(req *types.FileDownloadReq) (resp *types.FileDownloadResp, err error) {
	// TODO 在nginx层增加路由转发会更好
	return &types.FileDownloadResp{
		//Url: l.svcCtx.Oss.GetURLByPath(req.Path),
	}, nil
}
