package dictionary

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/logic/dictionary/word"
	"english-study/internal/utils"
	"fmt"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateExampleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewGenerateExampleLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *GenerateExampleLogic {
	return &GenerateExampleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *GenerateExampleLogic) GenerateExample(req *types.GenerateExampleReq) (resp *types.GenerateExampleResp, err error) {

	// 1. 查单词
	wp, err := l.svcCtx.Model.GetWordPos(l.ctx, req.WordPosId, &l.ui.ID)
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询单词失败").WithCause(err)
	}

	example, err := word.NewWordInfo(l.svcCtx, l.ui.ID).GenerateExample(l.ctx, wp.Word, wp.Pos, 1, req.Translation)
	if err != nil {
		return nil, err
	}
	if len(example) == 0 {
		return nil, errors.ErrorResourceGenerateError("生成例句失败").WithCause(fmt.Errorf("例句数量为0"))
	}

	return &types.GenerateExampleResp{
		Example: *example[0],
	}, nil
}
