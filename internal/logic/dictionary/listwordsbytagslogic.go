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

type ListWordsByTagsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewListWordsByTagsLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *ListWordsByTagsLogic {
	return &ListWordsByTagsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

// ListWordsByTags 按标签 AND 筛选: 返回"同时拥有全部所选标签"的词/短语 (服务端实时查 word_tags, 避免前端快照过期/关联错乱).
func (l *ListWordsByTagsLogic) ListWordsByTags(req *types.ListWordsByTagsReq) (resp *types.ListWordsByTagsResp, err error) {
	resp = &types.ListWordsByTagsResp{Data: make([]*types.TaggedWord, 0)}
	uid := l.ui.ID
	wt := req.WordType
	if wt != types.WordTypeWord && wt != types.WordTypePhrase {
		return nil, errors.ErrorRequestParamError("word_type 不合法")
	}
	// 标签去重; 为空时不筛选(前端无选中标签时本就不调用)
	uniq := make(map[uint]struct{}, len(req.TagIDs))
	tagIDs := make([]uint, 0, len(req.TagIDs))
	for _, id := range req.TagIDs {
		if id == 0 {
			continue
		}
		if _, ok := uniq[id]; ok {
			continue
		}
		uniq[id] = struct{}{}
		tagIDs = append(tagIDs, id)
	}
	if len(tagIDs) == 0 {
		return resp, nil
	}

	m := l.svcCtx.Model

	// 1. AND: 同时拥有全部所选标签的 word_id (GROUP BY ... HAVING COUNT(DISTINCT tag_id) = N)
	var ids []uint
	if err = m.DB.WithContext(l.ctx).
		Table("word_tags").
		Where("user_id = ? AND word_type = ? AND tag_id IN ?", uid, wt, tagIDs).
		Group("word_id").
		Having("COUNT(DISTINCT tag_id) = ?", len(tagIDs)).
		Pluck("word_id", &ids).Error; err != nil {
		return nil, errors.ErrorDatabaseQueryError("按标签筛选失败").WithCause(err)
	}
	if len(ids) == 0 {
		return resp, nil
	}

	// 2. 取词条文本 (per-user 表)
	type wrow struct {
		ID   uint
		Word string
	}
	var rows []wrow
	if wt == types.WordTypeWord {
		err = m.DB.WithContext(l.ctx).Table((&bean.Word{}).UserTableName(&uid)).
			Select("id, word").Where("id IN ?", ids).Order("word asc").Scan(&rows).Error
	} else {
		err = m.DB.WithContext(l.ctx).Table((&bean.WordPhrase{}).UserTableName(&uid)).
			Select("id, phrase AS word").Where("id IN ?", ids).Order("phrase asc").Scan(&rows).Error
	}
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询词条失败").WithCause(err)
	}
	if len(rows) == 0 {
		return resp, nil
	}

	// 3. 取这些词条的全部标签 (用于展示 chips)
	matchedIDs := make([]uint, 0, len(rows))
	for _, r := range rows {
		matchedIDs = append(matchedIDs, r.ID)
	}
	type trow struct {
		WordID uint
		TagID  uint
		Name   string
		Style  string
	}
	var trows []trow
	if err = m.DB.WithContext(l.ctx).
		Table("word_tags AS wt").
		Select("wt.word_id AS word_id, t.id AS tag_id, t.tag AS name, t.style AS style").
		Joins("JOIN tags t ON t.id = wt.tag_id").
		Where("wt.user_id = ? AND wt.word_type = ? AND wt.word_id IN ?", uid, wt, matchedIDs).
		Order("wt.word_id, t.id").
		Scan(&trows).Error; err != nil {
		return nil, errors.ErrorDatabaseQueryError("查询标签失败").WithCause(err)
	}
	tagsByWord := make(map[uint][]*types.SimpleTagInfo, len(rows))
	for _, t := range trows {
		tagsByWord[t.WordID] = append(tagsByWord[t.WordID], &types.SimpleTagInfo{ID: t.TagID, Name: t.Name, Style: t.Style})
	}

	for _, r := range rows {
		resp.Data = append(resp.Data, &types.TaggedWord{
			ID:   r.ID,
			Word: r.Word,
			Tags: tagsByWord[r.ID],
		})
	}
	return resp, nil
}
