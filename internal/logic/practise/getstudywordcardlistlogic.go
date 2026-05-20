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

type GetStudyWordCardListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewGetStudyWordCardListLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *GetStudyWordCardListLogic {
	return &GetStudyWordCardListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *GetStudyWordCardListLogic) GetStudyWordCardList(req *types.GetWordCardListReq) (resp *types.GetWordCardListResp, err error) {

	resp = &types.GetWordCardListResp{}

	wsg := l.svcCtx.Model.Gen.WordStatus
	find := wsg.WithContext(l.ctx).Where(
		wsg.UserID.Eq(l.ui.ID),
		wsg.Status.Eq(types.WordStatusStudy),
	).Order(wsg.CreatedAt.Desc())
	if req.WordType != 0 {
		find = find.Where(wsg.WordType.Eq(req.WordType))
	}
	var wss []*bean.WordStatus
	if !req.Random {
		// 按创建时间倒序
		wss, err = find.Limit(req.Count).Offset(0).Find()
		if err != nil {
			return nil, errors.ErrorDatabaseQueryError("查询学习中的单词失败").WithCause(err)
		}
	} else {
		wss, err = getRandomWordStatus(find, req.Count)
		if err != nil {
			return nil, err
		}
	}

	for _, v := range wss {
		wc, err := wordStatusToWordCard(l.ctx, l.svcCtx.Model, v)
		if err != nil {
			return nil, errors.ErrorRequestParamError("查询学习中的单词失败").WithCause(err)
		}
		resp.Data = append(resp.Data, wc)
	}

	return
}
