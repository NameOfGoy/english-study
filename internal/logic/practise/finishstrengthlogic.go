package practise

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/model/bean"
	"english-study/internal/utils"
	"time"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	err = l.svcCtx.Model.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		var ws bean.WordStatus
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND word_id = ? AND word_type = ? AND status = ?",
				l.ui.ID, req.WordID, req.WordType, types.WordStatusStrengthen).
			Take(&ws).Error; e != nil {
			return errors.ErrorDatabaseQueryError("查询单词状态失败").WithCause(e)
		}
		if ws.StudyTime.After(getTodayZeroTime()) {
			return errors.ErrorRequestParamError("今天已经强化过了")
		}
		statusBefore := ws.Status
		switch req.Operation {
		case 1:
			ws.Times++
		case 2:
			ws.Times = 0
		}
		newStatus, ferr := statusTransferFSM(&ws, req.Operation, &Rule{StrengthenTimes: 3})
		if ferr != nil {
			return errors.ErrorDatabaseUpdateError("状态转换失败").WithCause(ferr)
		}
		ws.Status = newStatus
		if statusBefore != ws.Status {
			ws.Times = 0
			ws.Interval = 1
			ws.Repetitions = 0
			ws.NextReviewAt = time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)
		}
		ws.StudyTime = time.Now()

		if e := tx.Model(&bean.WordStatus{}).Where("id = ?", ws.ID).Updates(map[string]interface{}{
			"status":         ws.Status,
			"times":          ws.Times,
			"interval":       ws.Interval,
			"repetitions":    ws.Repetitions,
			"next_review_at": ws.NextReviewAt,
			"study_time":     ws.StudyTime,
		}).Error; e != nil {
			return errors.ErrorDatabaseUpdateError("更新单词状态失败").WithCause(e)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &types.FinishStrengthResp{}, nil
}
