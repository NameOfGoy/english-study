package dictionary

import (
	"context"
	"strings"

	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchAddStardictLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewBatchAddStardictLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *BatchAddStardictLogic {
	return &BatchAddStardictLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

// maxBatchAddItems 单次请求上限, 防止前端误传 10k 项
const maxBatchAddItems = 500

// batchAddConcurrency 后台 goroutine 并发上限, 防止 LLM/OSS 连接耗尽
const batchAddConcurrency = 10

func (l *BatchAddStardictLogic) BatchAddStardict(req *types.BatchAddStardictReq) (*types.BatchAddStardictResp, error) {
	if len(req.Items) == 0 {
		return &types.BatchAddStardictResp{Submitted: 0}, nil
	}
	if len(req.Items) > maxBatchAddItems {
		req.Items = req.Items[:maxBatchAddItems]
	}

	// 用 buffered channel 当信号量, 限制后台 goroutine 总并发数
	sem := make(chan struct{}, batchAddConcurrency)
	submitted := 0
	for _, item := range req.Items {
		sw := strings.TrimSpace(item.Sw)
		if sw == "" {
			continue
		}
		submitted++
		wordType := item.WordType
		ui := l.ui
		svcCtx := l.svcCtx
		go func(sw string, wt int) {
			sem <- struct{}{}        // 阻塞直到有空槽
			defer func() { <-sem }() // 完成后释放
			bgCtx := context.Background()
			if wt == 2 {
				logic := NewAddWordPhraseLogic(bgCtx, svcCtx, ui)
				if _, err := logic.AddWordPhrase(&types.AddWordPhraseReq{
					Phrase:            sw,
					IsGeneratePicture: false,
				}); err != nil {
					logx.Errorf("批量添加短语失败, phrase: %s, err: %v", sw, err)
				}
			} else {
				logic := NewAddWordLogic(bgCtx, svcCtx, ui)
				if _, err := logic.AddWord(&types.AddWordReq{
					Word:              sw,
					IsGeneratePicture: false,
				}); err != nil {
					logx.Errorf("批量添加单词失败, word: %s, err: %v", sw, err)
				}
			}
		}(sw, wordType)
	}

	return &types.BatchAddStardictResp{Submitted: submitted}, nil
}
