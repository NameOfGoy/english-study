package dictionary

import (
	"context"
	"english-study/internal/utils"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetWordStatusListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewGetWordStatusListLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *GetWordStatusListLogic {
	return &GetWordStatusListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *GetWordStatusListLogic) GetWordStatusList(req *types.GetWordStatusListReq) (resp *types.GetWordStatusListResp, err error) {

	resp = &types.GetWordStatusListResp{}

	wsg := l.svcCtx.Model.Gen.WordStatus
	find := wsg.WithContext(l.ctx)

	// 仅查询当前用户
	find = find.Where(wsg.UserID.Eq(l.ui.ID))

	if len(req.WordID) > 0 {
		find = find.Where(wsg.WordID.In(req.WordID...))
	}
	// 过滤单词类型
	if req.WordType != 0 {
		find = find.Where(wsg.WordType.Eq(req.WordType))
	}

	// 暂不做分页

	// 查询
	wss, err := find.Find()
	if err != nil {
		return nil, err
	}

	for _, w := range wss {
		resp.Data = append(resp.Data, &types.WordStatus{
			ID:       w.ID,
			WordID:   w.WordID,
			WordType: w.WordType,
			Status:   w.Status,
			Times:    w.Times,
			Weight:   w.Weight,
		})
	}

	return
}
