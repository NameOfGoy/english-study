package user

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/utils"
	e "errors"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
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

	// 强制只能更新当前登录用户自己的资料，禁止通过路径 ID 越权改他人
	uid := l.ui.ID
	// 1. 查用户
	user, err := l.svcCtx.Model.Gen.User.Where(l.svcCtx.Model.Gen.User.ID.Eq(uid)).WithContext(l.ctx).First()
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询用户失败").WithCause(err)
	}
	// 2. 更新
	user.Username = req.UserInfo.Name
	// 账号: openid 自动注册的随机账号允许用户改, 但只能改 1 次
	if req.UserInfo.Account != "" && req.UserInfo.Account != user.Account {
		if user.AccountRenamed {
			return nil, errors.ErrorRequestParamError("账号只能修改一次")
		}
		ug := l.svcCtx.Model.Gen.User
		if _, qerr := ug.WithContext(l.ctx).Where(ug.Account.Eq(req.UserInfo.Account)).First(); qerr == nil {
			return nil, errors.ErrorAccountExistError("账号已被占用")
		} else if !e.Is(qerr, gorm.ErrRecordNotFound) {
			return nil, errors.ErrorDatabaseQueryError("查询账号失败").WithCause(qerr)
		}
		user.Account = req.UserInfo.Account
		user.AccountRenamed = true
	}
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
