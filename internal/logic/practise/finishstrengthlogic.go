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

type FinishStrengthLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewFinishStrengthLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *FinishStrengthLogic {
	return &FinishStrengthLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *FinishStrengthLogic) FinishStrength(req *types.FinishStrengthReq) (resp *types.FinishStrengthResp, err error) {

	// 查出单词状态
	wsg := l.svcCtx.Model.Gen.WordStatus
	ws, err := wsg.WithContext(l.ctx).Where(
		wsg.UserID.Eq(l.ui.ID),
		wsg.WordID.Eq(req.WordID),
		wsg.WordType.Eq(req.WordType),
		wsg.Status.Eq(types.WordStatusStrengthen),
	).Take()
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询单词状态失败").WithCause(err)
	}
	// 判断学习时间
	if ws.StudyTime.After(getTodayZeroTime()) {
		return nil, errors.ErrorRequestParamError("今天已经强化过了")
	}
	status := ws.Status
	switch req.Operation {
	case 1:
		ws.Times++ // 学习次数+1
	case 2:
		ws.Times = 0 // 要连续完成3次才能强化成功
	}
	ws.Status = statusTransferFSM(ws, req.Operation, &Rule{StrengthenTimes: 3}) // 状态转换
	if status != ws.Status {
		ws.Times = 0 // 状态变更，学习次数重置为0
		// Strengthen→Review: 保留EF，间隔重置为1天重新开始
		ws.Interval = 1
		ws.Repetitions = 0
		ws.NextReviewAt = time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)
	}
	ws.StudyTime = time.Now()

	_, err = wsg.WithContext(l.ctx).Where(wsg.ID.Eq(ws.ID)).Updates(ws)
	if err != nil {
		return nil, errors.ErrorDatabaseUpdateError("更新单词状态失败").WithCause(err)
	}

	return &types.FinishStrengthResp{}, nil
}
