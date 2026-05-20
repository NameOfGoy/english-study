package dictionary

import (
	"context"
	"english-study/internal/utils"

	"english-study/internal/svc"
	"english-study/internal/types"

	"english-study/internal/errors"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateWordPhraseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewUpdateWordPhraseLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *UpdateWordPhraseLogic {
	return &UpdateWordPhraseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *UpdateWordPhraseLogic) UpdateWordPhrase(req *types.UpdateWordPhraseReq) (resp *types.UpdateWordPhraseResp, err error) {
	// todo: add your logic here and delete this line

	return nil, errors.ErrorNotSupportError("更新短语暂不支持")
}
