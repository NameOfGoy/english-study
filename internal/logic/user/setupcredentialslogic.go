package user

import (
	"context"
	e "errors"
	"strings"

	"english-study/internal/errors"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type SetupCredentialsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewSetupCredentialsLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *SetupCredentialsLogic {
	return &SetupCredentialsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

// SetupCredentials 微信自动注册用户首次设置真实账号+密码. 仅占位态(密码为 wxonly: 前缀)允许, 一次性.
// 不验旧密码(占位密码本就不可用); 校验账号唯一 + 新密码强度; 设置后把账号也锁成"已改"(AccountRenamed).
func (l *SetupCredentialsLogic) SetupCredentials(req *types.SetupCredentialsReq) (resp *types.SetupCredentialsResp, err error) {
	account := strings.TrimSpace(req.Account)
	if account == "" || req.Password == "" {
		return nil, errors.ErrorRequestParamError("参数错误")
	}
	// wx_ 前缀是自动注册账号的保留前缀, 不允许用户设成这个(否则前端"占位"判断会失准)
	if strings.HasPrefix(account, "wx_") {
		return nil, errors.ErrorRequestParamError("账号不能以 wx_ 开头")
	}
	if perr := utils.ValidatePasswordStrength(req.Password); perr != nil {
		return nil, errors.ErrorRequestParamError(perr.Error())
	}

	uid := l.ui.ID
	ug := l.svcCtx.Model.Gen.User

	user, err := ug.WithContext(l.ctx).Where(ug.ID.Eq(uid)).First()
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询用户失败").WithCause(err)
	}
	// 仅占位密码(微信自动注册)才允许首次设置; 已设过真实密码的拒绝, 防覆盖/重复
	if !strings.HasPrefix(user.Password, "wxonly:") {
		return nil, errors.ErrorRequestParamError("账号密码已设置, 无需重复设置")
	}

	// 账号唯一(排除自己)
	if existing, qerr := ug.WithContext(l.ctx).Where(ug.Account.Eq(account)).First(); qerr == nil {
		if existing.ID != uid {
			return nil, errors.ErrorAccountExistError("账号已被占用")
		}
	} else if !e.Is(qerr, gorm.ErrRecordNotFound) {
		return nil, errors.ErrorDatabaseQueryError("查询账号失败").WithCause(qerr)
	}

	hashed, herr := utils.HashPassword(req.Password)
	if herr != nil {
		return nil, errors.ErrorDatabaseUpdateError("密码处理失败").WithCause(herr)
	}

	// 一次性设置: 账号 + 密码 + 标记账号已改(AccountRenamed, 之后普通编辑也不能再改账号)。
	// AccountRenamed 列未在 gen DAO 暴露字段表达式, 故沿用 updateuserinfo 的 bean.Updates 写法。
	// 注意: Updates(bean) 是"整 bean 非零字段"语义, 会把 user 上所有非零字段(含 WxOpenID/Username/Role 等)
	// 一并回写其当前值(幂等无副作用); 这里真正变更的是 account/password/AccountRenamed 三项(均非零必落库)。
	user.Account = account
	user.Password = hashed
	user.AccountRenamed = true
	if _, uerr := ug.WithContext(l.ctx).Where(ug.ID.Eq(uid)).Updates(user); uerr != nil {
		return nil, errors.ErrorDatabaseUpdateError("设置失败").WithCause(uerr)
	}

	return &types.SetupCredentialsResp{}, nil
}
