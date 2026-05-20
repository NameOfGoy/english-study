package tag

import (
	"context"
	"strings"

	"english-study/internal/errors"
	"english-study/internal/model/bean"
	"english-study/internal/utils"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddTagLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewAddTagLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *AddTagLogic {
	return &AddTagLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *AddTagLogic) AddTag(req *types.AddTagReq) (*types.AddTagResp, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.ErrorRequestParamError("标签名称不能为空")
	}
	if len([]rune(name)) > 20 {
		return nil, errors.ErrorRequestParamError("标签名称不能超过20字")
	}

	tg := l.svcCtx.Model.Gen.Tag

	// 重名检查：在自己的标签和默认标签里
	cnt, err := tg.WithContext(l.ctx).
		Where(tg.Tag.Eq(name)).
		Where(tg.WithContext(l.ctx).Where(tg.UserID.Eq(l.ui.ID)).Or(tg.UserID.Eq(0))).
		Count()
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询标签重名失败").WithCause(err)
	}
	if cnt > 0 {
		return nil, errors.ErrorRequestParamError("标签名已存在")
	}

	if err := tg.WithContext(l.ctx).Create(&bean.Tag{
		Tag:    name,
		Style:  req.Style,
		UserID: l.ui.ID,
	}); err != nil {
		return nil, errors.ErrorDatabaseInsertError("创建标签失败").WithCause(err)
	}

	return &types.AddTagResp{}, nil
}
