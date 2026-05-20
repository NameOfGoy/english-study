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

type AddWordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

// 字典模块/单词添加
func NewAddWordLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *AddWordLogic {
	return &AddWordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *AddWordLogic) AddWord(req *types.AddWordReq) (resp *types.AddWordResp, err error) {
	// 先检查用户是否已经添加过该单词
	if req.Word != "" {
		existing, e := l.svcCtx.Model.GetWordWithPosByWord(l.ctx, req.Word, &l.ui.ID)
		if e == nil && existing != nil {
			return nil, errors.ErrorRequestParamError("该单词已在你的词典中")
		}
		if e != nil && !stderrors.Is(e, gorm.ErrRecordNotFound) {
			return nil, errors.ErrorDatabaseQueryError("查询单词失败").WithCause(e)
		}
	}

	// 单词表新增单词
	wi := word.NewWordInfo(l.svcCtx, l.ui.ID)

	// 获取用户级的单词基本信息
	mainWord, err := wi.GetCustomizedWordInfo(l.ctx, &types.Word{
		Word:       req.Word,
		UKPhonetic: req.UKPhonetic,
		UKAudio:    req.UKAudio,
		USPhonetic: req.USPhonetic,
		USAudio:    req.USAudio,
		Pos:        req.Pos,
	})
	if err != nil {
		return nil, errors.ErrorWordNotExistError("单词不存在").WithCause(err)
	}

	// 生成图片
	if req.IsGeneratePicture {
		for _, wp := range mainWord.Pos {
			if link, err := wi.GeneratePicture(l.ctx, mainWord.Word, wp.Pos); err != nil {
				logx.Errorf("生成图片失败, word: %s, pos: %d, err: %v", mainWord.Word, wp.Pos, err)
			} else {
				wp.Picture = link
			}
		}
	}

	// 存入用户单词表
	err = wi.IncreaseWord(l.ctx, mainWord, &l.ui.ID)
	if err != nil {
		return nil, errors.ErrorDatabaseInsertError("新增单词失败").WithCause(err)
	}

	return &types.AddWordResp{}, nil
}
