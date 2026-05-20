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

type FinishStudyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewFinishStudyLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *FinishStudyLogic {
	return &FinishStudyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *FinishStudyLogic) FinishStudy(req *types.FinishStudyReq) (resp *types.FinishStudyResp, err error) {

	// 查出单词状态
	wsg := l.svcCtx.Model.Gen.WordStatus
	ws, err := wsg.WithContext(l.ctx).Where(
		wsg.UserID.Eq(l.ui.ID),
		wsg.WordID.Eq(req.WordID),
		wsg.WordType.Eq(req.WordType),
		wsg.Status.Eq(types.WordStatusStudy),
	).Take()
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询单词状态失败").WithCause(err)
	}
	status := ws.Status
	ws.Times++                                // 学习次数+1
	ws.Status = statusTransferFSM(ws, 1, nil) // 状态转换
	if status != ws.Status {
		ws.Times = 0 // 状态变更，学习次数重置为0
		// 进入Review: 初始化SRS参数
		ws.EaseFactor = 2.5
		ws.Interval = 1
		ws.Repetitions = 0
		ws.NextReviewAt = time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour) // 明天
	}
	ws.StudyTime = time.Now()

	_, err = wsg.WithContext(l.ctx).Where(wsg.ID.Eq(ws.ID)).Updates(ws)
	if err != nil {
		return nil, errors.ErrorDatabaseUpdateError("更新单词状态失败").WithCause(err)
	}

	return &types.FinishStudyResp{}, nil
}
