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

type DeleteWordPictureLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

// 字典模块/词性图片删除
func NewDeleteWordPictureLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *DeleteWordPictureLogic {
	return &DeleteWordPictureLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *DeleteWordPictureLogic) DeleteWordPicture(req *types.DeleteWordPictureReq) (resp *types.DeleteWordPictureResp, err error) {
	err = l.svcCtx.Model.DB.
		Table((&bean.WordPos{}).UserTableName(&l.ui.ID)).
		Where("id = ?", req.WordPosId).WithContext(l.ctx).
		Updates(map[string]interface{}{
			"picture": "",
		}).Error
	if err != nil {
		return nil, errors.ErrorDatabaseUpdateError("删除单词图片失败").WithCause(err)
	}
	return &types.DeleteWordPictureResp{}, nil
}
