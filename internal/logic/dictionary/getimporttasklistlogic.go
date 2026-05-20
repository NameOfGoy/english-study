package dictionary

import (
	"context"
	"strings"
	"time"

	"english-study/internal/errors"
	"english-study/internal/model/bean"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetImportTaskListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewGetImportTaskListLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *GetImportTaskListLogic {
	return &GetImportTaskListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *GetImportTaskListLogic) GetImportTaskList(req *types.GetImportTaskListReq) (*types.GetImportTaskListResp, error) {
	start, end, err := resolveTaskTimeRange(req)
	if err != nil {
		return nil, err
	}

	q := l.svcCtx.Model.DB.WithContext(l.ctx).
		Where("user_id = ?", l.ui.ID).
		Order("created_at DESC")
	if !start.IsZero() {
		q = q.Where("created_at >= ?", start)
	}
	if !end.IsZero() {
		q = q.Where("created_at < ?", end)
	}

	var tasks []bean.ImportTask
	if err := q.Find(&tasks).Error; err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询导入任务失败").WithCause(err)
	}

	items := make([]types.ImportTaskItem, 0, len(tasks))
	for _, t := range tasks {
		items = append(items, types.ImportTaskItem{
			Id:           t.ID,
			FileName:     t.FileName,
			Status:       t.Status,
			Total:        t.Total,
			Current:      t.Current,
			CurrentWord:  t.CurrentWord,
			SuccessCount: t.SuccessCount,
			FailCount:    t.FailCount,
			FailWords:    t.FailWords,
			CreatedAt:    t.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.GetImportTaskListResp{
		Tasks: items,
	}, nil
}

// resolveTaskTimeRange 解析时间范围。优先级:
//  1. 显式 start_date / end_date
//  2. days > 0
//  3. 默认最近 3 天
func resolveTaskTimeRange(req *types.GetImportTaskListReq) (start, end time.Time, err error) {
	startStr := strings.TrimSpace(req.StartDate)
	endStr := strings.TrimSpace(req.EndDate)

	if startStr != "" || endStr != "" {
		if startStr != "" {
			start, err = time.ParseInLocation("2006-01-02", startStr, time.Local)
			if err != nil {
				return start, end, errors.ErrorRequestParamError("start_date 格式应为 YYYY-MM-DD").WithCause(err)
			}
		}
		if endStr != "" {
			endTime, e := time.ParseInLocation("2006-01-02", endStr, time.Local)
			if e != nil {
				return start, end, errors.ErrorRequestParamError("end_date 格式应为 YYYY-MM-DD").WithCause(e)
			}
			end = endTime.Add(24 * time.Hour) // 含当天 → 上界取次日 00:00
		}
		return
	}

	days := req.Days
	if days <= 0 {
		days = 3
	}
	if days > 365 {
		days = 365
	}
	start = time.Now().AddDate(0, 0, -days).Truncate(24 * time.Hour)
	return
}
