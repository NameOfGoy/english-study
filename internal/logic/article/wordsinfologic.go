package article

import (
	"context"

	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type WordsInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewWordsInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *WordsInfoLogic {
	return &WordsInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

// WordsInfo 按词条文本批量查简要信息(词性/释义/音标); 词条不存在则 Found=false 留空.
func (l *WordsInfoLogic) WordsInfo(req *types.WordsInfoReq) (resp *types.WordsInfoResp, err error) {
	m := l.svcCtx.Model
	uid := l.ui.ID

	data := make([]*types.WordBriefItem, 0, len(req.Words))
	for _, w := range req.Words {
		b := m.GetWordBrief(l.ctx, uid, w.Word, w.Type)
		data = append(data, &types.WordBriefItem{
			Word:     w.Word,
			Type:     w.Type,
			PosLabel: b.PosLabel,
			Meaning:  b.Meaning,
			Senses:   toArticleSenses(b.Senses),
			Phonetic: b.Phonetic,
			Found:    b.Found,
		})
	}
	return &types.WordsInfoResp{Data: data}, nil
}
