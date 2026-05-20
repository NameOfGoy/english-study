package dictionary

import (
	"context"
	"english-study/internal/model/bean"
	"english-study/internal/utils"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	es "errors"
)

type UpdateWordStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewUpdateWordStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *UpdateWordStatusLogic {
	return &UpdateWordStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *UpdateWordStatusLogic) UpdateWordStatus(req *types.UpdateWordStatusReq) (resp *types.UpdateWordStatusResp, err error) {

	sg := l.svcCtx.Model.Gen.WordStatus
	// 如果不存在, 则创建
	if _, err = sg.Where(sg.WordID.Eq(req.WordID), sg.WordType.Eq(req.WordType), sg.UserID.Eq(l.ui.ID)).WithContext(l.ctx).Take(); err != nil {
		if !es.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		err = sg.WithContext(l.ctx).Create(&bean.WordStatus{
			WordID:   req.WordID,
			WordType: req.WordType,
			Status:   req.Status,
			UserID:   l.ui.ID,
		})
		if err != nil {
			return nil, err
		}
	}

	// 更新状态
	_, err = sg.WithContext(l.ctx).Where(sg.WordID.Eq(req.WordID), sg.WordType.Eq(req.WordType), sg.UserID.Eq(l.ui.ID)).Updates(map[string]interface{}{
		"status": req.Status,
	})
	if err != nil {
		return nil, err
	}

	return &types.UpdateWordStatusResp{}, nil
}
