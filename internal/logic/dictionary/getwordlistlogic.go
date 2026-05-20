package dictionary

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/model/bean"
	"english-study/internal/utils"
	"fmt"

	"english-study/internal/svc"
	"english-study/internal/types"

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
		if _, ok := l.svcCtx.Model.Gen.Word.GetFieldByName(req.OrderBy); !ok {
			return nil, errors.ErrorRequestParamError("排序字段不存在")
		}
		sort := "asc"
		if req.Sort == "desc" {
			sort = req.Sort
		}
		find = find.Order(fmt.Sprintf("%s %s", req.OrderBy, sort))
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
