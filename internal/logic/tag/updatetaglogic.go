package tag

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/utils"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTagLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewUpdateTagLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *UpdateTagLogic {
	return &UpdateTagLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *UpdateTagLogic) UpdateTag(req *types.UpdateTagReq) (resp *types.UpdateTagResp, err error) {
	tg := l.svcCtx.Model.Gen.Tag

	// 先取目标 tag, 校验归属
	target, err := tg.WithContext(l.ctx).Where(tg.ID.Eq(req.ID)).First()
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("标签不存在").WithCause(err)
	}
	if !canMutateTag(l.ui, target.UserID) {
		return nil, errors.ErrorPermissionError("无权操作该标签")
	}

	// 更新标签 (沿用原归属, 不允许通过 update 把用户标签改成系统标签或反之)
	_, err = tg.WithContext(l.ctx).Where(tg.ID.Eq(req.ID)).Updates(map[string]any{
		"tag":   req.Name,
		"style": req.Style,
	})
	if err != nil {
		return nil, errors.ErrorDatabaseUpdateError("更新标签失败").WithCause(err)
	}
	return &types.UpdateTagResp{}, nil
}

// canMutateTag 判断当前用户能否改 / 删一个 ownerID 拥有的标签
//   - ownerID == 0  (系统标签)  → 仅超管
//   - ownerID == ui.ID         → 允许
//   - 其它                     → 拒绝
func canMutateTag(ui *utils.UserInfo, ownerID uint) bool {
	if ownerID == 0 {
		return ui.IsAdmin()
	}
	return ui != nil && ownerID == ui.ID
}
