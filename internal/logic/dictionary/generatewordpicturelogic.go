package dictionary

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/logic/dictionary/word"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateWordPictureLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

// 字典模块/词性图片生成
func NewGenerateWordPictureLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *GenerateWordPictureLogic {
	return &GenerateWordPictureLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *GenerateWordPictureLogic) GenerateWordPicture(req *types.GenerateWordPictureReq) (resp *types.GenerateWordPictureResp, err error) {

	// 1. 查单词
	wp, err := l.svcCtx.Model.GetWordPos(l.ctx, req.WordPosId, &l.ui.ID)
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询单词失败").WithCause(err)
	}

	// 2. 生成图片
	link, err := word.NewWordInfo(l.svcCtx, l.ui.ID).GeneratePicture(l.ctx, wp.Word, wp.Pos)
	if err != nil {
		logx.Errorf("生成图片失败, wordPos: %v, err: %v", wp, err)
		return nil, errors.ErrorResourceGenerateError("生成图片失败").WithCause(err)
	}

	return &types.GenerateWordPictureResp{
		Link: utils.ToOssUri(types.OssBucket, link),
	}, nil
}
