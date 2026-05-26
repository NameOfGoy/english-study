package tag

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/model/bean"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTagListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewGetTagListLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *GetTagListLogic {
	return &GetTagListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *GetTagListLogic) GetTagList(req *types.GetTagListReq) (resp *types.GetTagListResp, err error) {

	resp = &types.GetTagListResp{}

	tg := l.svcCtx.Model.Gen.Tag

	// 系统标签 (user_id=0)
	defaultFind := tg.WithContext(l.ctx).Where(tg.UserID.Eq(0))
	defaultTotal, err := defaultFind.Count()
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询默认标签总数失败").WithCause(err)
	}
	defaultTags, err := defaultFind.Find()
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询默认标签失败").WithCause(err)
	}

	// 用户自己的标签 (user_id = ui.ID 且不是 0; 否则超管自己就是 user_id=0, 会和系统重复)
	var (
		tags  []*bean.Tag
		total int64
	)
	if l.ui.ID != 0 {
		find := tg.WithContext(l.ctx).Where(tg.UserID.Eq(l.ui.ID))
		total, err = find.Count()
		if err != nil {
			return nil, errors.ErrorDatabaseQueryError("查询用户标签总数失败").WithCause(err)
		}
		tags, err = find.Find()
		if err != nil {
			return nil, errors.ErrorDatabaseQueryError("查询用户标签失败").WithCause(err)
		}
	}
	// 合并系统标签和用户标签
	tagList := append(defaultTags, tags...)

	for _, tag := range tagList {
		resp.Data = append(resp.Data, &types.Tag{
			ID:       tag.ID,
			Name:     tag.Tag,
			Style:    tag.Style,
			IsSystem: tag.UserID == 0,
		})
	}
	resp.TotalCount = total + defaultTotal

	return
}
