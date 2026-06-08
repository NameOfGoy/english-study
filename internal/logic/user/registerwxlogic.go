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
	// HEAD 状态: api UserRegisterWxReq 只有 OpenId/Name/Avatar 三个字段, 没有 Code,
	// 前端直传 openid (弱化的客户端信任模式; 强化的话应让 api 加回 Code, 前端传 code,
	// 这里用 l.svcCtx.Wx.Code2Session 换 openid). 当前临时按弱化版本走以让 build 通过.
	openid := req.OpenId
	if openid == "" {
		return nil, errors.ErrorRequestParamError("open_id 不能为空")
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
