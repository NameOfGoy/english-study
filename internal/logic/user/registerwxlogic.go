package user

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/model/bean"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"
	e "errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type RegisterWxLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 用户模块/微信注册用户
func NewRegisterWxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterWxLogic {
	return &RegisterWxLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// RegisterWx 微信注册新账号: 用户在"关联/注册"弹窗选了"注册新账号"后, 前端用新的 wx.login code 调本接口。
// 服务端 code2session 换 openid → 自动注册一个随机账号(无手机号/邮箱, 名字默认"微信用户", 允许改名1次) → 签 token。
func (l *RegisterWxLogic) RegisterWx(req *types.UserRegisterWxReq) (resp *types.UserRegisterWxResp, err error) {
	// 1. code → openid
	openid, _, err := l.svcCtx.Wx.Code2Session(req.Code)
	if err != nil {
		return nil, errors.ErrorWxAuthError("微信登录失败").WithCause(err)
	}

	// 2. openid 已存在则幂等返回(直接登录), 否则自动注册
	ug := l.svcCtx.Model.Gen.User
	user, err := ug.Where(ug.WxOpenID.Eq(openid)).WithContext(l.ctx).First()
	if err != nil {
		if !e.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrorDatabaseQueryError("查询用户失败").WithCause(err)
		}
		// Account/WxOpenID/Password 均 not-null+unique: 给唯一随机账号 + 占位密码(不可用于密码登录)
		newUser := &bean.User{
			Account:  "wx_" + strings.ReplaceAll(uuid.NewString(), "-", "")[:16],
			WxOpenID: openid,
			Username: "微信用户",
			Password: "wxonly:" + uuid.NewString(),
		}
		if cerr := l.svcCtx.Model.CreateUser(l.ctx, newUser); cerr != nil {
			return nil, errors.ErrorDatabaseInsertError("微信自动注册失败").WithCause(cerr)
		}
		masked := openid
		if len(masked) > 6 {
			masked = masked[:6] + "***"
		}
		logx.WithContext(l.ctx).Infof("微信注册: 自动注册新用户 uid=%d openid_prefix=%s", newUser.ID, masked)
		user = newUser
	}

	// 3. 签 token
	token, err := utils.GenerateToken(l.svcCtx.Config.Auth.AccessSecret, time.Now().Unix(), l.svcCtx.Config.Auth.AccessExpire, user.ID, user.Username, user.Role)
	if err != nil {
		return nil, errors.ErrorTokenGenerateError("生成token失败").WithCause(err)
	}
	avatar := user.Avatar
	if avatar != "" {
		avatar = utils.ToOssUri(types.OssBucket, avatar)
	}
	return &types.UserRegisterWxResp{
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
