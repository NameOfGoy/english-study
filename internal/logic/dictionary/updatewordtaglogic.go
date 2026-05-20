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

type UpdateWordTagLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewUpdateWordTagLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *UpdateWordTagLogic {
	return &UpdateWordTagLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *UpdateWordTagLogic) UpdateWordTag(req *types.UpdateWordTagReq) (resp *types.UpdateWordTagResp, err error) {

	// 开事务
	tx := l.svcCtx.Model.Gen.Begin()
	wgTx := tx.WordTag
	defer func() {
		if err != nil {
			if txe := tx.Rollback(); txe != nil {
				l.Logger.Errorf("rollback tx failed: %v", txe)
			}
		} else {
			if txe := tx.Commit(); txe != nil {
				l.Logger.Errorf("commit tx failed: %v", txe)
			}
		}
	}()

	// 先删除原来的
	if _, err = wgTx.WithContext(l.ctx).Where(wgTx.WordID.Eq(req.WordID), wgTx.WordType.Eq(req.WordType), wgTx.UserID.Eq(l.ui.ID)).Delete(); err != nil {
		return nil, errors.ErrorDatabaseDeleteError("删除单词标签失败").WithCause(err)
	}
	// 插入新的
	for _, tag := range req.Tags {
		if err = wgTx.WithContext(l.ctx).Create(&bean.WordTag{
			WordID:   req.WordID,
			WordType: req.WordType,
			TagID:    tag,
			UserID:   l.ui.ID,
		}); err != nil {
			return nil, errors.ErrorDatabaseInsertError("创建单词标签失败").WithCause(err)
		}
	}

	return &types.UpdateWordTagResp{}, nil
}
