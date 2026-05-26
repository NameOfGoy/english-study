package user

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/utils"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

// 用户模块/获取用户信息
func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *GetUserInfoLogic) GetUserInfo(req *types.GetUserInfoReq) (resp *types.GetUserInfoResp, err error) {

	// 强制只能读取当前登录用户自己的资料，禁止通过路径 ID 越权读他人
	_ = req
	uid := l.ui.ID
	// 1. 查用户
	user, err := l.svcCtx.Model.Gen.User.Where(l.svcCtx.Model.Gen.User.ID.Eq(uid)).WithContext(l.ctx).First()
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询用户失败").WithCause(err)
	}
	// 2. 转换
	resp = &types.GetUserInfoResp{
		UserInfo: types.UserInfo{
			ID:      user.ID,
			Name:    user.Username,
			Account: user.Account,
			Phone:   user.Phone,
			Email:   user.Email,
			Avatar:  utils.ToOssUri(types.OssBucket, user.Avatar),
			Role:    user.Role,
		},
	}
	return resp, nil
}
