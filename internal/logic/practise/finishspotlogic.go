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

type FinishSpotLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewFinishSpotLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *FinishSpotLogic {
	return &FinishSpotLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *FinishSpotLogic) FinishSpot(req *types.FinishSpotReq) (resp *types.FinishSpotResp, err error) {

	// 查出单词状态
	wsg := l.svcCtx.Model.Gen.WordStatus
	ws, err := wsg.WithContext(l.ctx).Where(
		wsg.UserID.Eq(l.ui.ID),
		wsg.WordID.Eq(req.WordID),
		wsg.WordType.Eq(req.WordType),
		wsg.Status.Eq(types.WordStatusFinish),
	).Take()
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询单词状态失败").WithCause(err)
	}
	status := ws.Status
	ws.Status = statusTransferFSM(ws, req.Operation, &Rule{}) // 状态转换
	if status != ws.Status {
		ws.Times = 0 // 状态变更，学习次数重置为0
		// Finish→Strengthen: 重置SRS，间隔=1天
		ws.Interval = 1
		ws.Repetitions = 0
		ws.NextReviewAt = time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)
	}
	ws.StudyTime = time.Now()

	_, err = wsg.WithContext(l.ctx).Where(wsg.ID.Eq(ws.ID)).Updates(ws)
	if err != nil {
		return nil, errors.ErrorDatabaseUpdateError("更新单词状态失败").WithCause(err)
	}

	return &types.FinishSpotResp{}, nil
}
