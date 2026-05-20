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

type UpdateWordPictureLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

// 字典模块/词性图片更新
func NewUpdateWordPictureLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *UpdateWordPictureLogic {
	return &UpdateWordPictureLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *UpdateWordPictureLogic) UpdateWordPicture(req *types.UpdateWordPictureReq) (resp *types.UpdateWordPictureResp, err error) {

	err = l.svcCtx.Model.DB.
		Table((&bean.WordPos{}).UserTableName(&l.ui.ID)).
		Where("id = ?", req.WordPosId).WithContext(l.ctx).
		Updates(map[string]interface{}{
			"picture": utils.ToOssPath(types.OssBucket, req.Picture),
		}).Error
	if err != nil {
		return nil, errors.ErrorDatabaseUpdateError("更新单词失败").WithCause(err)
	}
	return &types.UpdateWordPictureResp{}, nil
}
