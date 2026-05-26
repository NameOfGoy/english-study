package dictionary

import (
	"context"
	"english-study/internal/utils"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetWordTagListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewGetWordTagListLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *GetWordTagListLogic {
	return &GetWordTagListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *GetWordTagListLogic) GetWordTagList(req *types.GetWordTagListReq) (resp *types.GetWordTagListResp, err error) {

	resp = &types.GetWordTagListResp{}

	wtg := l.svcCtx.Model.Gen.WordTag
	tg := l.svcCtx.Model.Gen.Tag
	find := wtg.WithContext(l.ctx)

	// 支持根据 WordID 和 WordType 过滤
	if len(req.WordID) > 0 {
		find = find.Where(wtg.WordID.In(req.WordID...))
	}
	if req.WordType != 0 {
		find = find.Where(wtg.WordType.Eq(req.WordType))
	}

	// 暂不做分页

	// 用扁平结构接 JOIN 结果, 避免 GORM Scan 到嵌入指针时不自动分配导致字段全 0
	type Result struct {
		ID       uint   `gorm:"column:id"`
		WordID   uint   `gorm:"column:word_id"`
		WordType int    `gorm:"column:word_type"`
		TagID    uint   `gorm:"column:tag_id"`
		Name     string `gorm:"column:tag"`
		Style    string `gorm:"column:style"`
	}
	var wts []Result
	err = find.Select(
		wtg.ID,
		wtg.WordID,
		wtg.WordType,
		wtg.TagID,
		tg.Tag,
		tg.Style,
	).LeftJoin(tg, wtg.TagID.EqCol(tg.ID)).
		Where(wtg.UserID.Eq(l.ui.ID)).
		Order(wtg.WordID, wtg.TagID).
		Scan(&wts)
	if err != nil {
		return nil, err
	}

	for _, w := range wts {
		resp.Data = append(resp.Data, &types.WordTag{
			ID:       w.ID,
			WordID:   w.WordID,
			WordType: w.WordType,
			TagID:    w.TagID,
			Name:     w.Name,
			Style:    w.Style,
		})
	}

	return
}
