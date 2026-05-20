package dictionary

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/types"
	"english-study/internal/utils"

	"english-study/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDictionaryCountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewGetDictionaryCountLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *GetDictionaryCountLogic {
	return &GetDictionaryCountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *GetDictionaryCountLogic) GetDictionaryCount() (resp *types.GetDictionaryCountResp, err error) {
	resp = &types.GetDictionaryCountResp{}

	wsg := l.svcCtx.Model.Gen.WordStatus

	wordCount, err := wsg.WithContext(l.ctx).Where(
		wsg.UserID.Eq(l.ui.ID),
		wsg.WordType.Eq(types.WordTypeWord),
	).Count()
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询单词数量失败").WithCause(err)
	}

	phraseCount, err := wsg.WithContext(l.ctx).Where(
		wsg.UserID.Eq(l.ui.ID),
		wsg.WordType.Eq(types.WordTypePhrase),
	).Count()
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询短语数量失败").WithCause(err)
	}

	resp.WordCount = int(wordCount)
	resp.PhraseCount = int(phraseCount)
	resp.TotalCount = int(wordCount + phraseCount)

	return resp, nil
}
