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

	// 系统标签需要超管权限; 鉴权依据是 JWT 解出来的 role, 不读请求体的其它字段
	ownerID := l.ui.ID
	if req.IsSystem {
		if err := utils.RequireAdmin(l.ui); err != nil {
			return nil, errors.ErrorPermissionError("仅超管可创建系统标签")
		}
		ownerID = 0
	}

	tg := l.svcCtx.Model.Gen.Tag

	// 重名检查: 同名查 (owner + 系统标签); 系统标签 owner=0 时, "系统标签" 那部分等价于自身, 不必再 union
	q := tg.WithContext(l.ctx).Where(tg.Tag.Eq(name))
	if ownerID == 0 {
		q = q.Where(tg.UserID.Eq(0))
	} else {
		q = q.Where(tg.WithContext(l.ctx).Where(tg.UserID.Eq(ownerID)).Or(tg.UserID.Eq(0)))
	}
	cnt, err := q.Count()
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询标签重名失败").WithCause(err)
	}
	if cnt > 0 {
		return nil, errors.ErrorRequestParamError("标签名已存在")
	}

	if err := tg.WithContext(l.ctx).Create(&bean.Tag{
		Tag:    name,
		Style:  req.Style,
		UserID: ownerID,
	}); err != nil {
		return nil, errors.ErrorDatabaseInsertError("创建标签失败").WithCause(err)
	}

	return &types.AddTagResp{}, nil
}
