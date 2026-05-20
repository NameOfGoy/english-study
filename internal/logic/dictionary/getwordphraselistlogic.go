package dictionary

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/model/bean"
	"english-study/internal/utils"
	"fmt"
	"strings"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetWordPhraseListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewGetWordPhraseListLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *GetWordPhraseListLogic {
	return &GetWordPhraseListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *GetWordPhraseListLogic) GetWordPhraseList(req *types.GetWordPhraseListReq) (resp *types.GetWordPhraseListResp, err error) {

	find := l.svcCtx.Model.DB.Table((&bean.WordPhrase{}).UserTableName(&l.ui.ID)).WithContext(l.ctx)

	// 先不处理word phrase ids

	// 处理短语前缀
	if req.PhrasePrefix != "" {
		find = find.Where("lower(phrase) like ?", strings.ToLower(req.PhrasePrefix)+"%")
	}
	// 处理翻译
	if req.Translation != "" {
		find = find.Where("translation like ?", "%"+req.Translation+"%")
	}

	var count int64
	err = find.Count(&count).Error
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询短语列表失败").WithCause(err)
	}

	if req.Offset != 0 {
		find = find.Offset(req.Offset)
	}
	if req.Limit != 0 {
		find = find.Limit(req.Limit)
	}
	if req.OrderBy != "" {
		if _, ok := l.svcCtx.Model.Gen.Word.GetFieldByName(req.OrderBy); !ok {
			return nil, errors.ErrorRequestParamError("排序字段不存在")
		}
		sort := "asc"
		if req.Sort == "desc" {
			sort = req.Sort
		}
		find = find.Order(fmt.Sprintf("%s %s", req.OrderBy, sort))
	}

	var list []*bean.WordPhrase
	err = find.Find(&list).Error
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询短语列表失败").WithCause(err)
	}

	resp = &types.GetWordPhraseListResp{Data: make([]*types.SimpleWord, 0)}
	resp.TotalCount = count

	for _, v := range list {
		resp.Data = append(resp.Data, &types.SimpleWord{
			ID:   v.ID,
			Word: v.Phrase,
		})
	}

	return
}
