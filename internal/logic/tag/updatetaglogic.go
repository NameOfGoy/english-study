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
	if !canMutateTag(l.ui, target.IsSystem, target.UserID) {
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

// canMutateTag 判断当前用户能否改 / 删一个标签
//   - isSystem == true (系统标签) → 仅超管
//   - 否则 ownerID == ui.ID       → 允许
//   - 其它                       → 拒绝
//
// 注意: 用 isSystem 而非 ownerID==0 判系统标签, 因为真实用户(sssadmin)的 user_id 也可能是 0,
// 其私有标签(is_system=false, user_id=0)必须只能本人操作, 不能被当成系统标签放给所有超管.
func canMutateTag(ui *utils.UserInfo, isSystem bool, ownerID uint) bool {
	if ui == nil {
		return false
	}
	if isSystem {
		return ui.IsAdmin()
	}
	return ownerID == ui.ID
}
