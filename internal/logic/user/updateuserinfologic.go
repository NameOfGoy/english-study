package user

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/utils"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

// 用户模块/更新用户信息
func NewUpdateUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *UpdateUserInfoLogic {
	return &UpdateUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *UpdateUserInfoLogic) UpdateUserInfo(req *types.UpdateUserInfoReq) (resp *types.UpdateUserInfoResp, err error) {

	uid, err := req.GetUintId()
	if err != nil {
		uid = l.ui.ID
	}
	// 1. 查用户
	user, err := l.svcCtx.Model.Gen.User.Where(l.svcCtx.Model.Gen.User.ID.Eq(uid)).WithContext(l.ctx).First()
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询用户失败").WithCause(err)
	}
	// 2. 更新
	user.Username = req.UserInfo.Name
	user.Account = req.UserInfo.Account
	user.Phone = req.UserInfo.Phone
	user.Email = req.UserInfo.Email
	user.Avatar = utils.ToOssPath(types.OssBucket, req.UserInfo.Avatar)
	// 3. 保存
	_, err = l.svcCtx.Model.Gen.User.WithContext(l.ctx).Where(l.svcCtx.Model.Gen.User.ID.Eq(uid)).Updates(user)
	if err != nil {
		return nil, errors.ErrorDatabaseUpdateError("更新用户失败").WithCause(err)
	}
	return &types.UpdateUserInfoResp{}, nil
}
