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

type DeletePhrasePictureLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewDeletePhrasePictureLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *DeletePhrasePictureLogic {
	return &DeletePhrasePictureLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *DeletePhrasePictureLogic) DeletePhrasePicture(req *types.DeletePhrasePictureReq) (resp *types.DeletePhrasePictureResp, err error) {
	err = l.svcCtx.Model.DB.
		Table((&bean.WordPhrase{}).UserTableName(&l.ui.ID)).
		Where("id = ?", req.ID).WithContext(l.ctx).
		Updates(map[string]interface{}{
			"picture": "",
		}).Error
	if err != nil {
		return nil, errors.ErrorDatabaseUpdateError("删除单词图片失败").WithCause(err)
	}
	return &types.DeletePhrasePictureResp{}, nil
}
