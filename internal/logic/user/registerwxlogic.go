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
	if req.OpenId == "" {
		return nil, errors.ErrorRequestParamError("openid不能为空")
	}
	// 1. 检查用户是否存在
	_, err = l.svcCtx.Model.Gen.User.Where(l.svcCtx.Model.Gen.User.WxOpenID.Eq(req.OpenId)).WithContext(l.ctx).First()
	if err == nil {
		return nil, errors.ErrorAccountExistError("用户已存在")
	}
	if !e.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.ErrorDatabaseQueryError("查询失败").WithCause(err)
	}

	// 2. 创建用户
	user := &bean.User{
		WxOpenID: req.OpenId,
		Username: req.Name,
		Avatar:   req.Avatar,
	}
	err = l.svcCtx.Model.CreateUser(l.ctx, user)
	if err != nil {
		return nil, errors.ErrorDatabaseInsertError("创建用户失败").WithCause(err)
	}

	return &types.UserRegisterWxResp{}, nil
}
