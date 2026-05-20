package dictionary

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/logic/dictionary/word"
	"english-study/internal/utils"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GeneratePhraseExampleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewGeneratePhraseExampleLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *GeneratePhraseExampleLogic {
	return &GeneratePhraseExampleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *GeneratePhraseExampleLogic) GeneratePhraseExample(req *types.GeneratePhraseExampleReq) (resp *types.GeneratePhraseExampleResp, err error) {
	// 1. 查单词
	wp, err := l.svcCtx.Model.GetWordPhraseById(l.ctx, req.ID, &l.ui.ID)
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询单词失败").WithCause(err)
	}

	// 2. 生成例句
	examples, err := word.NewPhraseInfo(l.svcCtx, l.ui.ID).GenerateExample(l.ctx, wp.Phrase, 1, "")
	if err != nil {
		logx.Errorf("生成例句失败, phrase: %v, err: %v", wp, err)
		return nil, errors.ErrorResourceGenerateError("生成图片失败").WithCause(err)
	}
	if len(examples) == 0 {
		return nil, errors.ErrorResourceGenerateError("生成例句失败: 未返回结果")
	}

	return &types.GeneratePhraseExampleResp{
		Example: *examples[0],
	}, nil
}
