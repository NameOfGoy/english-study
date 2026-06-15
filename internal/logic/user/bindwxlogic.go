package user

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"
	e "errors"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type BindWxLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBindWxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindWxLogic {
	return &BindWxLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// BindWx 微信关联已有账号: 用户选"关联已有账号"后输账号密码, 前端用新 wx.login code 调本接口。
// 验证账号密码(证明所有权) → 把 wxid 绑到该账号 → 签 token 登录。公开接口(用户此时尚未登录)。
func (l *BindWxLogic) BindWx(req *types.UserBindWxReq) (resp *types.UserBindWxResp, err error) {
	// 1. 校验账号密码
	if strings.TrimSpace(req.Account) == "" || req.Password == "" {
		return nil, errors.ErrorRequestParamError("账号或密码不能为空")
	}
	ug := l.svcCtx.Model.Gen.User
	user, err := ug.WithContext(l.ctx).Where(ug.Account.Eq(req.Account)).First()
	if err != nil {
		if e.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrorRequestParamError("账号或密码错误")
		}
		return nil, errors.ErrorDatabaseQueryError("查询用户失败").WithCause(err)
	}
	if ok, _ := utils.VerifyPassword(user.Password, req.Password); !ok {
		return nil, errors.ErrorRequestParamError("账号或密码错误")
	}

	// 2. code → openid
	openid, _, err := l.svcCtx.Wx.Code2Session(req.Code)
	if err != nil {
		return nil, errors.ErrorWxAuthError("微信登录失败").WithCause(err)
	}

	// 3. 绑定校验
	if user.WxOpenID != "" && !strings.HasPrefix(user.WxOpenID, "non-wx:") {
		// 该账号已绑过真实微信
		if user.WxOpenID != openid {
			return nil, errors.ErrorNotSupportError("该账号已绑定其他微信")
		}
		// 否则就是当前微信, 幂等放行
	} else {
		// 该 openid 不能已被别的账号占用
		if other, oerr := ug.WithContext(l.ctx).Where(ug.WxOpenID.Eq(openid)).First(); oerr == nil {
			if other.ID != user.ID {
				return nil, errors.ErrorNotSupportError("该微信已关联其他账号")
			}
		} else if !e.Is(oerr, gorm.ErrRecordNotFound) {
			return nil, errors.ErrorDatabaseQueryError("查询失败").WithCause(oerr)
		}
		// 绑定 wxid 到该账号
		if _, uerr := ug.WithContext(l.ctx).Where(ug.ID.Eq(user.ID)).Update(ug.WxOpenID, openid); uerr != nil {
			return nil, errors.ErrorDatabaseUpdateError("绑定微信失败").WithCause(uerr)
		}
		user.WxOpenID = openid
	}

	// 4. 签 token 登录
	token, err := utils.GenerateToken(l.svcCtx.Config.Auth.AccessSecret, time.Now().Unix(), l.svcCtx.Config.Auth.AccessExpire, user.ID, user.Username, user.Role)
	if err != nil {
		return nil, errors.ErrorTokenGenerateError("生成token失败").WithCause(err)
	}
	avatar := user.Avatar
	if avatar != "" {
		avatar = utils.ToOssUri(types.OssBucket, avatar)
	}
	return &types.UserBindWxResp{
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
