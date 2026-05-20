package dictionary

import (
	"context"
	stderrors "errors"
	"english-study/internal/errors"
	"english-study/internal/logic/dictionary/word"
	"english-study/internal/utils"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type AddWordPhraseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewAddWordPhraseLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *AddWordPhraseLogic {
	return &AddWordPhraseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *AddWordPhraseLogic) AddWordPhrase(req *types.AddWordPhraseReq) (resp *types.AddWordPhraseResp, err error) {
	// 先检查用户是否已经添加过该短语
	if req.Phrase != "" {
		existing, e := l.svcCtx.Model.GetWordPhraseByPhrase(l.ctx, req.Phrase, &l.ui.ID)
		if e == nil && existing != nil {
			return nil, errors.ErrorRequestParamError("该短语已在你的词典中")
		}
		if e != nil && !stderrors.Is(e, gorm.ErrRecordNotFound) {
			return nil, errors.ErrorDatabaseQueryError("查询短语失败").WithCause(e)
		}
	}

	wi := word.NewPhraseInfo(l.svcCtx, l.ui.ID)

	// 取得用户级别的短语表
	wordPhrase, err := wi.GetCustomizedPhraseInfo(l.ctx, &types.WordPhrase{
		Phrase: req.Phrase,
	})
	if err != nil {
		return nil, err
	}

	// 生成图片
	if req.IsGeneratePicture {
		wordPhrase.Picture, err = wi.GeneratePicture(l.ctx, req.Phrase)
		if err != nil {
			return nil, err
		}
	}

	// 存入用户短语表
	err = wi.IncreasePhrase(l.ctx, wordPhrase, &l.ui.ID)
	if err != nil {
		return nil, err
	}

	return &types.AddWordPhraseResp{}, nil
}
