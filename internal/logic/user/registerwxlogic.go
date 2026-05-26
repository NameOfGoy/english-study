package user

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/model/bean"
	e "errors"

	"gorm.io/gorm"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
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

func (l *RegisterWxLogic) RegisterWx(req *types.UserRegisterWxReq) (resp *types.UserRegisterWxResp, err error) {
	if req.Code == "" {
		return nil, errors.ErrorRequestParamError("code不能为空")
	}
	// 1. 用 code 向微信换取真实的 openid（防止客户端伪造 openid）
	openid, _, err := l.svcCtx.Wx.Code2Session(req.Code)
	if err != nil {
		return nil, errors.ErrorWxAuthError("微信换取openid失败").WithCause(err)
	}
	if openid == "" {
		return nil, errors.ErrorWxAuthError("微信返回空 openid")
	}

	// 2. 检查用户是否存在
	_, err = l.svcCtx.Model.Gen.User.Where(l.svcCtx.Model.Gen.User.WxOpenID.Eq(openid)).WithContext(l.ctx).First()
	if err == nil {
		return nil, errors.ErrorAccountExistError("用户已存在")
	}
	if !e.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.ErrorDatabaseQueryError("查询失败").WithCause(err)
	}

	// 3. 创建用户
	user := &bean.User{
		WxOpenID: openid,
		Username: req.Name,
		Avatar:   req.Avatar,
	}
	err = l.svcCtx.Model.CreateUser(l.ctx, user)
	if err != nil {
		return nil, errors.ErrorDatabaseInsertError("创建用户失败").WithCause(err)
	}

	return &types.UserRegisterWxResp{}, nil
}
