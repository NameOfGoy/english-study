package share

import (
	"context"
	"time"

	"english-study/internal/errors"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateShareLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewGenerateShareLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *GenerateShareLogic {
	return &GenerateShareLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *GenerateShareLogic) GenerateShare(req *types.GenerateShareReq) (*types.GenerateShareResp, error) {
	if req.ShareType != ShareTypeAll && req.ShareType != ShareTypeByTag {
		return nil, errors.ErrorRequestParamError("share_type 非法")
	}
	if req.WordType < 0 || req.WordType > 2 {
		return nil, errors.ErrorRequestParamError("word_type 非法")
	}
	if req.ShareType == ShareTypeByTag && len(req.TagIDs) == 0 {
		return nil, errors.ErrorRequestParamError("按标签分享必须选择至少一个标签")
	}
	if len(req.TagIDs) > 255 {
		return nil, errors.ErrorRequestParamError("标签数量超过上限")
	}

	tagIDs := make([]uint32, len(req.TagIDs))
	for i, id := range req.TagIDs {
		tagIDs[i] = uint32(id)
	}

	expireAt := time.Now().Add(TokenTTL).Unix()

	payload := &SharePayload{
		UserID:    uint32(l.ui.ID),
		ShareType: uint8(req.ShareType),
		WordType:  uint8(req.WordType),
		TagIDs:    tagIDs,
		ExpireTs:  uint32(expireAt),
		Nonce:     NewNonce(),
	}

	token, err := EncodeToken(payload, l.svcCtx.Config.Auth.AccessSecret)
	if err != nil {
		return nil, errors.ErrorRequestParamError("生成分享码失败").WithCause(err)
	}

	return &types.GenerateShareResp{
		Token:     token,
		ExpiresAt: expireAt,
	}, nil
}
