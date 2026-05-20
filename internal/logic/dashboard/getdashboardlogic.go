package dashboard

import (
	"context"
	"time"

	"english-study/internal/errors"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDashboardLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewGetDashboardLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *GetDashboardLogic {
	return &GetDashboardLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

type dashboardAgg struct {
	TotalWords        int `gorm:"column:total_words"`
	StudyCount        int `gorm:"column:study_count"`
	ReviewCount       int `gorm:"column:review_count"`
	StrengthenCount   int `gorm:"column:strengthen_count"`
	SpotCount         int `gorm:"column:spot_count"`
	FinishedWords     int `gorm:"column:finished_words"`
	TodayStudied      int `gorm:"column:today_studied"`
	TodayReviewed     int `gorm:"column:today_reviewed"`
	TodayStrengthened int `gorm:"column:today_strengthened"`
}

func (l *GetDashboardLogic) GetDashboard() (*types.GetDashboardResp, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.Add(24 * time.Hour)

	var agg dashboardAgg
	err := l.svcCtx.Model.DB.WithContext(l.ctx).
		Table("word_statuses").
		Select(`
			COUNT(*) AS total_words,
			COUNT(*) FILTER (WHERE status = 1) AS study_count,
			COUNT(*) FILTER (WHERE status = 2 AND (next_review_at IS NULL OR next_review_at <= ?)) AS review_count,
			COUNT(*) FILTER (WHERE status = 3 AND (next_review_at IS NULL OR next_review_at <= ?)) AS strengthen_count,
			COUNT(*) FILTER (WHERE status = 4) AS spot_count,
			COUNT(*) FILTER (WHERE status = 4) AS finished_words,
			COUNT(*) FILTER (WHERE status = 2 AND study_time >= ? AND study_time < ?) AS today_studied,
			COUNT(*) FILTER (WHERE status = 2 AND study_time >= ? AND study_time < ? AND times > 0) AS today_reviewed,
			COUNT(*) FILTER (WHERE status = 3 AND study_time >= ? AND study_time < ?) AS today_strengthened
		`, now, now, today, tomorrow, today, tomorrow, today, tomorrow).
		Where("user_id = ?", l.ui.ID).
		Scan(&agg).Error
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询看板数据失败").WithCause(err)
	}

	progressRate := 0.0
	if agg.TotalWords > 0 {
		progressRate = float64(agg.FinishedWords) / float64(agg.TotalWords)
	}

	return &types.GetDashboardResp{
		Data: types.DashboardData{
			StudyCount:        agg.StudyCount,
			ReviewCount:       agg.ReviewCount,
			StrengthenCount:   agg.StrengthenCount,
			SpotCount:         agg.SpotCount,
			TodayStudied:      agg.TodayStudied,
			TodayReviewed:     agg.TodayReviewed,
			TodayStrengthened: agg.TodayStrengthened,
			TotalWords:        agg.TotalWords,
			FinishedWords:     agg.FinishedWords,
			ProgressRate:      progressRate,
		},
	}, nil
}
