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
	// 整段 read-modify-write 必须在事务里, 用 SELECT FOR UPDATE 防双击/重试导致 SRS 字段被双倍递增
	err = l.svcCtx.Model.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		var ws bean.WordStatus
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND word_id = ? AND word_type = ? AND status = ?",
				l.ui.ID, req.WordID, req.WordType, types.WordStatusReview).
			Take(&ws).Error; e != nil {
			return errors.ErrorDatabaseQueryError("查询单词状态失败").WithCause(e)
		}

		// SRS计算
		quality := req.Quality
		if quality < 1 || quality > 5 {
			quality = 4
		}
		statusBefore := ws.Status

		if req.Operation == 1 { // 成功
			srs := SM2Calculate(ws.EaseFactor, ws.Interval, ws.Repetitions, quality)
			ws.EaseFactor = srs.EaseFactor
			ws.Interval = srs.Interval
			ws.Repetitions = srs.Repetitions
			ws.NextReviewAt = srs.NextReview
			ws.Times++
			newStatus, ferr := statusTransferFSM(&ws, req.Operation, nil)
			if ferr != nil {
				return errors.ErrorDatabaseUpdateError("状态转换失败").WithCause(ferr)
			}
			ws.Status = newStatus
		} else { // 失败 → Strengthen, 间隔重置
			ws.Interval = 1
			ws.Repetitions = 0
			ws.NextReviewAt = time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)
			ws.Times++
			newStatus, ferr := statusTransferFSM(&ws, req.Operation, nil)
			if ferr != nil {
				return errors.ErrorDatabaseUpdateError("状态转换失败").WithCause(ferr)
			}
			ws.Status = newStatus
		}

		if statusBefore != ws.Status {
			ws.Times = 0
		}
		ws.StudyTime = time.Now()

		if e := tx.Model(&bean.WordStatus{}).Where("id = ?", ws.ID).Updates(map[string]interface{}{
			"status":         ws.Status,
			"times":          ws.Times,
			"ease_factor":    ws.EaseFactor,
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
	return &types.FinishReviewResp{}, nil
}
