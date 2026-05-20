package user

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 用户模块/用户登录
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.UserLoginReq) (resp *types.UserLoginResp, err error) {

	// 校验账号密码是否正确
	ug := l.svcCtx.Model.Gen.User
	user, err := ug.WithContext(l.ctx).Where(ug.Account.Eq(req.Account)).First()
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("账号不存在").WithCause(err)
	}
	// 校验密码是否正确
	if user.Password != req.Password {
		return nil, errors.ErrorDatabaseQueryError("密码错误")
	}

	// 登录成功，返回token
	token, err := utils.GenerateToken(l.svcCtx.Config.Auth.AccessSecret, time.Now().Unix(), l.svcCtx.Config.Auth.AccessExpire, user.ID, user.Username)
	if err != nil {
		return nil, errors.ErrorTokenGenerateError("生成token失败").WithCause(err)
	}

	return &types.UserLoginResp{
		Token: token,
		UserInfo: types.UserInfo{
			ID:      user.ID,
			Name:    user.Username,
			Email:   user.Email,
			Avatar:  utils.ToOssUri(types.OssBucket, user.Avatar),
			Phone:   user.Phone,
			Account: user.Account,
		},
	}, nil
}
