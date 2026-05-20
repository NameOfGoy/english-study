package dictionary

import (
	"context"
	"english-study/internal/utils"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/openpgp/errors"
)

type DeleteWordPhraseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewDeleteWordPhraseLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *DeleteWordPhraseLogic {
	return &DeleteWordPhraseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *DeleteWordPhraseLogic) DeleteWordPhrase(req *types.DeleteWordPhraseReq) (resp *types.DeleteWordPhraseResp, err error) {

	return nil, errors.UnsupportedError("删除短语功能暂未实现")
}
