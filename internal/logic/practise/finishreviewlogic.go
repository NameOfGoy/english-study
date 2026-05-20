package practise

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/utils"
	"time"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FinishReviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewFinishReviewLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *FinishReviewLogic {
	return &FinishReviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *FinishReviewLogic) FinishReview(req *types.FinishReviewReq) (resp *types.FinishReviewResp, err error) {

	// 查出单词状态
	wsg := l.svcCtx.Model.Gen.WordStatus
	ws, err := wsg.WithContext(l.ctx).Where(
		wsg.UserID.Eq(l.ui.ID),
		wsg.WordID.Eq(req.WordID),
		wsg.WordType.Eq(req.WordType),
		wsg.Status.Eq(types.WordStatusReview),
	).Take()
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询单词状态失败").WithCause(err)
	}

	// SRS计算
	quality := req.Quality
	if quality < 1 || quality > 5 {
		quality = 4
	}

	status := ws.Status

	if req.Operation == 1 { // 成功
		srs := SM2Calculate(ws.EaseFactor, ws.Interval, ws.Repetitions, quality)
		ws.EaseFactor = srs.EaseFactor
		ws.Interval = srs.Interval
		ws.Repetitions = srs.Repetitions
		ws.NextReviewAt = srs.NextReview
		ws.Times++
		ws.Status = statusTransferFSM(ws, req.Operation, nil)
	} else { // 失败 → Strengthen, 间隔重置
		ws.Interval = 1
		ws.Repetitions = 0
		ws.NextReviewAt = time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)
		ws.Times++
		ws.Status = statusTransferFSM(ws, req.Operation, nil)
	}

	if status != ws.Status {
		ws.Times = 0
	}
	ws.StudyTime = time.Now()

	_, err = wsg.WithContext(l.ctx).Where(wsg.ID.Eq(ws.ID)).Updates(ws)
	if err != nil {
		return nil, errors.ErrorDatabaseUpdateError("更新单词状态失败").WithCause(err)
	}

	return &types.FinishReviewResp{}, nil
}
