package article

import (
	"context"
	"strings"

	"english-study/internal/errors"
	"english-study/internal/logic/wordselect"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type ArticleCandidatesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewArticleCandidatesLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *ArticleCandidatesLogic {
	return &ArticleCandidatesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

// candidateCap 自选候选词上限(避免大词库一次拉太多).
const candidateCap = 300

// ArticleCandidates 自选 Page3 候选词: 与即时生成共用同一条选词链(状态/类别/标签), 返回轻量卡片.
func (l *ArticleCandidatesLogic) ArticleCandidates(req *types.ArticleCandidatesReq) (resp *types.ArticleCandidatesResp, err error) {
	m := l.svcCtx.Model
	uid := l.ui.ID

	find, ferr := buildWordStatusChain(l.ctx, m, uid, req.Status, req.Category, req.TagIDs)
	if ferr != nil {
		return nil, errors.ErrorDatabaseQueryError("筛选候选词失败").WithCause(ferr)
	}
	rows, qerr := find.Limit(candidateCap).Find()
	if qerr != nil {
		return nil, errors.ErrorDatabaseQueryError("查询候选词失败").WithCause(qerr)
	}

	data := make([]*types.ArticleCandidate, 0, len(rows))
	for _, ws := range rows {
		card, cerr := wordselect.WordStatusToWordCard(l.ctx, m, ws)
		if cerr != nil {
			continue // 跳过查不到详情的(脏数据), 不影响整体
		}
		data = append(data, &types.ArticleCandidate{
			WordID:   card.ID,
			WordType: card.WordType,
			Word:     card.Word,
			PosLabel: firstPosLabel(card),
			Meaning:  strings.TrimSpace(card.Translation),
			Phonetic: pickPhonetic(card),
		})
	}
	return &types.ArticleCandidatesResp{Data: data}, nil
}
