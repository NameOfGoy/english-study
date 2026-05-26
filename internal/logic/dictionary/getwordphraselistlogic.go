package dictionary

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/model/bean"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"
	"strings"

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
		// 严格白名单（且按 WordPhrase 实际列名，不再借用 Gen.Word 的字段表）
		allowedOrderBy := map[string]string{
			"phrase":      "phrase",
			"translation": "translation",
			"created_at":  "created_at",
			"updated_at":  "updated_at",
			"id":          "id",
		}
		col, ok := allowedOrderBy[req.OrderBy]
		if !ok {
			return nil, errors.ErrorRequestParamError("排序字段不允许")
		}
		direction := "asc"
		if req.Sort == "desc" {
			direction = "desc"
		}
		find = find.Order(col + " " + direction)
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
