package share

import (
	"context"

	"english-study/internal/errors"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetShareWordDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewGetShareWordDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *GetShareWordDetailLogic {
	return &GetShareWordDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *GetShareWordDetailLogic) GetShareWordDetail(req *types.GetShareWordDetailReq) (*types.GetShareWordDetailResp, error) {
	payload, err := DecodeToken(req.Token, DeriveShareSecret(l.svcCtx.Config.Auth.AccessSecret))
	if err != nil {
		return nil, errors.ErrorRequestParamError("分享码无效或已过期").WithCause(err)
	}
	sourceUserID := uint(payload.UserID)

	// 校验 req.WordID 必须在 token 授权的范围内（防止持有任意一个分享 token 就能枚举对方所有单词）
	allowedWordIDs, allowedPhraseIDs, err := collectShareItemIDs(l.ctx, l.svcCtx, sourceUserID, payload)
	if err != nil {
		return nil, err
	}
	inSet := func(ids []uint, target uint) bool {
		for _, id := range ids {
			if id == target {
				return true
			}
		}
		return false
	}

	resp := &types.GetShareWordDetailResp{}

	if req.WordType == 1 {
		if !inSet(allowedWordIDs, req.WordID) {
			return nil, errors.ErrorRequestParamError("该单词不在分享范围内")
		}
		// 单词详情：复用 A 端的 word_user_{A} + word_pos_user_{A}
		word, err := l.svcCtx.Model.GetWordWithPosById(l.ctx, req.WordID, &sourceUserID)
		if err != nil {
			return nil, errors.ErrorDatabaseQueryError("查询A端单词详情失败").WithCause(err)
		}
		w := &types.Word{
			ID:         word.ID,
			Word:       word.Word,
			UKPhonetic: word.BritishPronunciation,
			UKAudio:    utils.ToOssUri(types.OssBucket, word.BritishPronunciationAudio),
			USPhonetic: word.AmericanPronunciation,
			USAudio:    utils.ToOssUri(types.OssBucket, word.AmericanPronunciationAudio),
			Pos:        make([]*types.WordPos, 0, len(word.Pos)),
		}
		for _, pos := range word.Pos {
			p := &types.WordPos{
				ID:          pos.ID,
				WordID:      pos.WordID,
				Word:        pos.Word,
				Pos:         pos.Pos,
				Translation: pos.Translation,
				Picture:     utils.ToOssUri(types.OssBucket, pos.Picture),
			}
			_ = p.ExampleObject(pos.Example)
			p.ExchangeObject(pos.Exchange)
			w.Pos = append(w.Pos, p)
		}
		resp.Word = w
	} else if req.WordType == 2 {
		if !inSet(allowedPhraseIDs, req.WordID) {
			return nil, errors.ErrorRequestParamError("该短语不在分享范围内")
		}
		// 短语详情：复用 A 端的 word_phrase_user_{A}
		ph, err := l.svcCtx.Model.GetWordPhraseById(l.ctx, req.WordID, &sourceUserID)
		if err != nil {
			return nil, errors.ErrorDatabaseQueryError("查询A端短语详情失败").WithCause(err)
		}
		p := &types.WordPhrase{
			ID:            ph.ID,
			Phrase:        ph.Phrase,
			Translation:   ph.Translation,
			Pronunciation: utils.ToOssUri(types.OssBucket, ph.Pronunciation),
			Picture:       utils.ToOssUri(types.OssBucket, ph.Picture),
		}
		_ = p.ExampleObject(ph.Example)
		resp.Phrase = p
	} else {
		return nil, errors.ErrorRequestParamError("word_type 非法")
	}

	return resp, nil
}
