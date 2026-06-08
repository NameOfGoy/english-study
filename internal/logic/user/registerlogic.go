package user

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/model/bean"
	"english-study/internal/utils"
	e "errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 用户模块/用户注册
func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.UserRegisterReq) (resp *types.UserRegisterResp, err error) {

	// 1. 校验参数
	if req.Account == "" || req.Password == "" || req.Name == "" {
		return nil, errors.ErrorRequestParamError("参数错误")
	}
	// 2. 校验账号是否存在
	_, err = l.svcCtx.Model.Gen.User.WithContext(l.ctx).Where(l.svcCtx.Model.Gen.User.Account.Eq(req.Account)).First()
	if err != nil {
		if e.Is(err, gorm.ErrRecordNotFound) {
			// 账号不存在，继续注册
		} else {
			return nil, errors.ErrorDatabaseQueryError("查询用户失败").WithCause(err)
		}
	} else {
		// 账号已存在
		return nil, errors.ErrorAccountExistError("账号已存在")
	}
	// 3. 创建用户（密码用 PBKDF2 哈希后存储）
	hashedPwd, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.ErrorDatabaseInsertError("密码处理失败").WithCause(err)
	}
	user := &bean.User{
		Account:  req.Account,
		WxOpenID: "non-wx:" + uuid.New().String(), // non-WeChat accounts need a unique placeholder
		Password: hashedPwd,
		Username: req.Name,
		Phone:    req.Phone,
		Email:    req.Email,
	}
	err = l.svcCtx.Model.CreateUser(l.ctx, user)
	if err != nil {
		return nil, errors.ErrorDatabaseInsertError("创建用户失败").WithCause(err)
	}
	// 4. 返回结果
	return &types.UserRegisterResp{
		CommonReply: types.CommonReply{
			Code: 0,
			Msg:  "注册成功",
		},
	}, nil
}
