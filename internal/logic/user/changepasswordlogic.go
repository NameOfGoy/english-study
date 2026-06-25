package user

import (
	"context"

	"english-study/internal/errors"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangePasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewChangePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *ChangePasswordLogic {
	return &ChangePasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

// ChangePassword 修改当前登录用户的密码: 校验原密码 → 校验新密码强度 → bcrypt 入库. 仅改自己, 不越权.
func (l *ChangePasswordLogic) ChangePassword(req *types.ChangePasswordReq) (resp *types.ChangePasswordResp, err error) {
	if req.OldPassword == "" || req.NewPassword == "" {
		return nil, errors.ErrorRequestParamError("参数错误")
	}

	uid := l.ui.ID
	ug := l.svcCtx.Model.Gen.User

	// 1. 查当前登录用户
	user, err := ug.WithContext(l.ctx).Where(ug.ID.Eq(uid)).First()
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询用户失败").WithCause(err)
	}

	// 2. 校验原密码(兼容历史存储格式)
	if ok, _ := utils.VerifyPassword(user.Password, req.OldPassword); !ok {
		return nil, errors.ErrorRequestParamError("原密码错误")
	}

	// 2.5 新密码不能与原密码相同(否则改了个寂寞还被强制重登)
	if req.NewPassword == req.OldPassword {
		return nil, errors.ErrorRequestParamError("新密码不能与原密码相同")
	}

	// 3. 新密码强度校验(与注册同一套规则)
	if perr := utils.ValidatePasswordStrength(req.NewPassword); perr != nil {
		return nil, errors.ErrorRequestParamError(perr.Error())
	}

	// 4. bcrypt 哈希后只更新 password 列
	hashed, herr := utils.HashPassword(req.NewPassword)
	if herr != nil {
		return nil, errors.ErrorDatabaseUpdateError("密码处理失败").WithCause(herr)
	}
	if _, uerr := ug.WithContext(l.ctx).Where(ug.ID.Eq(uid)).Update(ug.Password, hashed); uerr != nil {
		return nil, errors.ErrorDatabaseUpdateError("更新密码失败").WithCause(uerr)
	}

	return &types.ChangePasswordResp{}, nil
}
