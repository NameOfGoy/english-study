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

type GeneratePhrasePictureLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewGeneratePhrasePictureLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *GeneratePhrasePictureLogic {
	return &GeneratePhrasePictureLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *GeneratePhrasePictureLogic) GeneratePhrasePicture(req *types.GeneratePhrasePictureReq) (resp *types.GeneratePhrasePictureResp, err error) {
	// 1. 查单词
	wp, err := l.svcCtx.Model.GetWordPhraseById(l.ctx, req.ID, &l.ui.ID)
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询单词失败").WithCause(err)
	}

	// 2. 生成图片
	link, err := word.NewPhraseInfo(l.svcCtx, l.ui.ID).GeneratePicture(l.ctx, wp.Phrase)
	if err != nil {
		logx.Errorf("生成图片失败, phrase: %v, err: %v", wp, err)
		return nil, errors.ErrorResourceGenerateError("生成图片失败").WithCause(err)
	}

	return &types.GeneratePhrasePictureResp{
		Link: utils.ToOssUri(types.OssBucket, link),
	}, nil
}
