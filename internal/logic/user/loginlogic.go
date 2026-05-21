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

// dummyHash 用 bcrypt 当"假对照", 给账号不存在分支做对等耗时, 防止时序探测账号是否注册
// (实际硬编码任意密码生成的 hash 都行)
const dummyHash = "$2a$12$abcdefghijklmnopqrstuv1234567890abcdefghijklmnopqrstuvw"

func (l *LoginLogic) Login(req *types.UserLoginReq) (resp *types.UserLoginResp, err error) {

	ug := l.svcCtx.Model.Gen.User
	user, err := ug.WithContext(l.ctx).Where(ug.Account.Eq(req.Account)).First()
	if err != nil {
		// 账号不存在: 仍然走一次 VerifyPassword 消耗等量 CPU 时间, 防时序探测; 返回模糊错
		_, _ = utils.VerifyPassword(dummyHash, req.Password)
		return nil, errors.ErrorRequestParamError("账号或密码错误")
	}
	// 校验密码：兼容历史明文，并在登录成功后自动升级到 bcrypt
	ok, needUpgrade := utils.VerifyPassword(user.Password, req.Password)
	if !ok {
		// 同一文案, 不区分账号不存在 / 密码错, 防枚举
		return nil, errors.ErrorRequestParamError("账号或密码错误")
	}
	if needUpgrade {
		if hashed, hashErr := utils.HashPassword(req.Password); hashErr == nil {
			if _, upErr := ug.WithContext(l.ctx).Where(ug.ID.Eq(user.ID)).
				Updates(map[string]interface{}{"password": hashed}); upErr != nil {
				logx.WithContext(l.ctx).Errorf("升级用户 %d 密码哈希失败: %v", user.ID, upErr)
			}
		}
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
