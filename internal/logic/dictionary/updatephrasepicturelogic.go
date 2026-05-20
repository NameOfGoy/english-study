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

type UpdatePhrasePictureLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewUpdatePhrasePictureLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *UpdatePhrasePictureLogic {
	return &UpdatePhrasePictureLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *UpdatePhrasePictureLogic) UpdatePhrasePicture(req *types.UpdatePhrasePictureReq) (resp *types.UpdatePhrasePictureResp, err error) {
	err = l.svcCtx.Model.DB.
		Table((&bean.WordPhrase{}).UserTableName(&l.ui.ID)).
		Where("id = ?", req.ID).WithContext(l.ctx).
		Updates(map[string]interface{}{
			"picture": utils.ToOssPath(types.OssBucket, req.Link),
		}).Error
	if err != nil {
		return nil, errors.ErrorDatabaseUpdateError("更新单词失败").WithCause(err)
	}

	return &types.UpdatePhrasePictureResp{}, nil
}
