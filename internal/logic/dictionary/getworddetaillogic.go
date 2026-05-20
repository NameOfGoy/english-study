package dictionary

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/utils"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetWordDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

// 字典模块/单词详情
func NewGetWordDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *GetWordDetailLogic {
	return &GetWordDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *GetWordDetailLogic) GetWordDetail(req *types.GetWordDetailReq) (resp *types.GetWordDetailResp, err error) {

	word, err := l.svcCtx.Model.GetWordWithPosById(l.ctx, req.Id, &l.ui.ID)
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询单词详情失败").WithCause(err)
	}

	// 结构转换
	w := types.Word{
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
	return &types.GetWordDetailResp{Data: w}, nil
}
