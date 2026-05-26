package practise

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/model/bean"
	"english-study/internal/utils"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSpotWordCardListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewGetSpotWordCardListLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *GetSpotWordCardListLogic {
	return &GetSpotWordCardListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *GetSpotWordCardListLogic) GetSpotWordCardList(req *types.GetWordCardListReq) (resp *types.GetWordCardListResp, err error) {

	resp = &types.GetWordCardListResp{}

	wsg := l.svcCtx.Model.Gen.WordStatus
	find := wsg.WithContext(l.ctx).Where(
		wsg.UserID.Eq(l.ui.ID),
		wsg.Status.Eq(types.WordStatusFinish),
	).Order(wsg.CreatedAt.Desc())
	if req.WordType != 0 {
		find = find.Where(wsg.WordType.Eq(req.WordType))
	}
	find, err = applyTagFilter(l.ctx, l.svcCtx.Model, find, l.ui.ID, req.TagIDs)
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("应用标签筛选失败").WithCause(err)
	}

	var wss []*bean.WordStatus
	if !req.Random {
		// 按创建时间倒序
		wss, err = find.Limit(req.Count).Offset(0).Find()
		if err != nil {
			return nil, errors.ErrorDatabaseQueryError("查询抽查中的单词失败").WithCause(err)
		}
	} else {
		wss, err = getRandomWordStatus(find, req.Count)
		if err != nil {
			return nil, errors.ErrorDatabaseQueryError("查询抽查中的单词失败").WithCause(err)
		}
	}

	for _, v := range wss {
		wc, err := wordStatusToWordCard(l.ctx, l.svcCtx.Model, v)
		if err != nil {
			return nil, errors.ErrorRequestParamError("查询抽查中的单词失败").WithCause(err)
		}
		resp.Data = append(resp.Data, wc)
	}

	return
}
