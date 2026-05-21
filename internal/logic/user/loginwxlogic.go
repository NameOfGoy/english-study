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

func (l *LoginWxLogic) LoginWx(req *types.UserLoginWxReq) (resp *types.UserLoginWxResp, err error) {

	// 1. 调用微信接口，获取openid
	openid, _, err := l.svcCtx.Wx.Code2Session(req.Code)
	if err != nil {
		return nil, errors.ErrorWxAuthError("微信登录失败").WithCause(err)
	}

	// 2. 检查用户是否存在
	user, err := l.svcCtx.Model.Gen.User.Where(l.svcCtx.Model.Gen.User.WxOpenID.Eq(openid)).WithContext(l.ctx).First()
	if err != nil {
		if e.Is(err, gorm.ErrRecordNotFound) {
			// 不把完整 openid 写进 cause 或日志 (持久 PII), 只 log 前缀; 方便回查但日志泄漏后也不能精准定位个人
			masked := openid
			if len(masked) > 6 {
				masked = masked[:6] + "***"
			}
			logx.WithContext(l.ctx).Infof("微信登录: openid 未注册 openid_prefix=%s", masked)
			return nil, errors.ErrorAccountNotExistError("用户不存在")
		}
		return nil, errors.ErrorDatabaseQueryError("查询用户失败").WithCause(err)
	}

	// 生成token
	token, err := utils.GenerateToken(l.svcCtx.Config.Auth.AccessSecret, time.Now().Unix(), l.svcCtx.Config.Auth.AccessExpire, user.ID, user.Username)
	if err != nil {
		return nil, errors.ErrorTokenGenerateError("生成token失败").WithCause(err)
	}

	// 3. 返回token
	return &types.UserLoginWxResp{
		Token: token,
		UserInfo: types.UserInfo{
			ID:      user.ID,
			Name:    user.Username,
			Account: user.Account,
			Phone:   user.Phone,
			Email:   user.Email,
			Avatar:  user.Avatar,
		},
	}, nil
}
