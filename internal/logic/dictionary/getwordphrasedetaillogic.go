package dictionary

import (
	"context"
	"english-study/internal/utils"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetWordPhraseDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewGetWordPhraseDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *GetWordPhraseDetailLogic {
	return &GetWordPhraseDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *GetWordPhraseDetailLogic) GetWordPhraseDetail(req *types.GetWordPhraseDetailReq) (resp *types.GetWordPhraseDetailResp, err error) {
	w, err := l.svcCtx.Model.GetWordPhraseById(l.ctx, req.ID, &l.ui.ID)
	if err != nil {
		return nil, err
	}
	data := &types.WordPhrase{
		ID:            w.ID,
		Phrase:        w.Phrase,
		Translation:   w.Translation,
		Pronunciation: utils.ToOssUri(types.OssBucket, w.Pronunciation),
		Picture:       utils.ToOssUri(types.OssBucket, w.Picture),
	}
	err = data.ExampleObject(w.Example)
	if err != nil {
		return nil, err
	}
	return &types.GetWordPhraseDetailResp{
		Data: *data,
	}, nil
}
