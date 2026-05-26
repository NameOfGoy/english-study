package dictionary

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/model/bean"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetWordListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

// 字典模块/单词列表
func NewGetWordListLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *GetWordListLogic {
	return &GetWordListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *GetWordListLogic) GetWordList(req *types.GetWordListReq) (resp *types.GetWordListResp, err error) {

	// 查询
	find := l.svcCtx.Model.DB.Table((&bean.Word{}).UserTableName(&l.ui.ID)).WithContext(l.ctx)
	if req.WordPrefix != "" {
		find = find.Where("word like ?", req.WordPrefix+"%")
	}
	if req.Pos != 0 {
		return nil, errors.ErrorNotSupportError("词性查询暂不支持")
	}
	if req.Translation != "" {
		return nil, errors.ErrorNotSupportError("中文翻译查询暂不支持")
	}
	if req.Phonetic != "" {
		return nil, errors.ErrorNotSupportError("音标查询暂不支持")
	}
	var total int64
	err = find.Count(&total).Error
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询单词总数失败").WithCause(err)
	}
	if req.Offset != 0 {
		find = find.Offset(req.Offset)
	}
	if req.Limit != 0 {
		find = find.Limit(req.Limit)
	}
	if req.OrderBy != "" {
		// 严格白名单，禁止任何用户输入直接拼进 ORDER BY
		allowedOrderBy := map[string]string{
			"word":       "word",
			"created_at": "created_at",
			"updated_at": "updated_at",
			"id":         "id",
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
	var words []*bean.Word
	err = find.Find(&words).Error
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询单词列表失败").WithCause(err)
	}
	resp = &types.GetWordListResp{
		Data: make([]*types.SimpleWord, 0, len(words)),
	}
	resp.TotalCount = total
	for _, w := range words {
		resp.Data = append(resp.Data, &types.SimpleWord{
			ID:   w.ID,
			Word: w.Word,
		})
	}

	return resp, nil
}
