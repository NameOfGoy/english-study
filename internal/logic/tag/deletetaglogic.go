package tag

import (
	"context"

	"english-study/internal/errors"
	"english-study/internal/utils"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type DeleteTagLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewDeleteTagLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *DeleteTagLogic {
	return &DeleteTagLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *DeleteTagLogic) DeleteTag(req *types.DeleteTagReq) (*types.DeleteTagResp, error) {
	// 事务：级联删除 word_tags 关联 + 删除 tag
	// 只能删除自己的标签（默认标签 user_id=0 受保护）
	err := l.svcCtx.Model.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("word_tags").
			Where("tag_id = ? AND user_id = ?", req.ID, l.ui.ID).
			Delete(nil).Error; err != nil {
			return err
		}
		return tx.Table("tags").
			Where("id = ? AND user_id = ?", req.ID, l.ui.ID).
			Delete(nil).Error
	})
	if err != nil {
		return nil, errors.ErrorDatabaseDeleteError("删除标签失败").WithCause(err)
	}
	return &types.DeleteTagResp{}, nil
}
