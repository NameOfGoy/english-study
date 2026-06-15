package user

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"
	e "errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type LoginWxLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 用户模块/用户登录/微信登录
func NewLoginWxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginWxLogic {
	return &LoginWxLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// LoginWx 微信登录: 进入登录页自动调用。code2session 换 openid:
//   - openid 已关联账号 → 签 token 直接登录;
//   - openid 未关联 → 返回空 token(不在此自动注册), 由前端弹窗让用户选 关联已有账号 / 注册新账号。
func (l *LoginWxLogic) LoginWx(req *types.UserLoginWxReq) (resp *types.UserLoginWxResp, err error) {
	// 1. code → openid
	openid, _, err := l.svcCtx.Wx.Code2Session(req.Code)
	if err != nil {
		return nil, errors.ErrorWxAuthError("微信登录失败").WithCause(err)
	}

	// 2. 查该 openid 是否已关联账号
	ug := l.svcCtx.Model.Gen.User
	user, err := ug.Where(ug.WxOpenID.Eq(openid)).WithContext(l.ctx).First()
	if err != nil {
		if e.Is(err, gorm.ErrRecordNotFound) {
			// 未关联: 返回空 token(无错误), 前端据此弹窗。不在此自动注册。
			masked := openid
			if len(masked) > 6 {
				masked = masked[:6] + "***"
			}
			logx.WithContext(l.ctx).Infof("微信登录: openid 未关联账号 openid_prefix=%s", masked)
			return &types.UserLoginWxResp{}, nil
		}
		return nil, errors.ErrorDatabaseQueryError("查询用户失败").WithCause(err)
	}

	// 3. 已关联 → 签 token 登录
	token, err := utils.GenerateToken(l.svcCtx.Config.Auth.AccessSecret, time.Now().Unix(), l.svcCtx.Config.Auth.AccessExpire, user.ID, user.Username, user.Role)
	if err != nil {
		return nil, errors.ErrorTokenGenerateError("生成token失败").WithCause(err)
	}
	avatar := user.Avatar
	if avatar != "" {
		avatar = utils.ToOssUri(types.OssBucket, avatar)
	}
	return &types.UserLoginWxResp{
		Token: token,
		UserInfo: types.UserInfo{
			ID:      user.ID,
			Name:    user.Username,
			Account: user.Account,
			Phone:   user.Phone,
			Email:   user.Email,
			Avatar:  avatar,
			Role:    user.Role,
		},
	}, nil
}
