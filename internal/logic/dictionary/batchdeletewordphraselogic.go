package dictionary

import (
	"context"
	"fmt"

	"english-study/internal/errors"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// maxBatchDeletePhrase 单次批量删除短语上限
const maxBatchDeletePhrase = 500

type BatchDeleteWordPhraseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

// 字典模块/短语批量删除
func NewBatchDeleteWordPhraseLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *BatchDeleteWordPhraseLogic {
	return &BatchDeleteWordPhraseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

// BatchDeleteWordPhrase 批量删除当前用户的多个短语及级联数据(学习状态/标签关联), 单事务原子.
// 去重 + 过滤 0; 不存在/非本人的 id 自动忽略; 返回实际删除数. 与 DeleteWordPhrase 级联一致, 仅改为 IN.
func (l *BatchDeleteWordPhraseLogic) BatchDeleteWordPhrase(req *types.BatchDeleteWordPhraseReq) (resp *types.BatchDeleteWordPhraseResp, err error) {
	ids := dedupIDs(req.IDs)
	if len(ids) == 0 {
		return nil, errors.ErrorRequestParamError("未选择要删除的短语")
	}
	if len(ids) > maxBatchDeletePhrase {
		return nil, errors.ErrorRequestParamError("一次最多删除 500 个短语")
	}

	uid := l.ui.ID
	phraseTable := fmt.Sprintf("word_phrase_user_%d", uid)

	var deleted int64
	err = l.svcCtx.Model.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Table(phraseTable).Where("id IN ?", ids).Delete(nil)
		if res.Error != nil {
			return errors.ErrorDatabaseDeleteError("删除短语失败").WithCause(res.Error)
		}
		deleted = res.RowsAffected
		if deleted == 0 {
			return nil
		}
		if e := tx.Table("word_statuses").
			Where("user_id = ? AND word_type = ? AND word_id IN ?", uid, types.WordTypePhrase, ids).
			Delete(nil).Error; e != nil {
			return errors.ErrorDatabaseDeleteError("删除短语学习状态失败").WithCause(e)
		}
		if e := tx.Table("word_tags").
			Where("user_id = ? AND word_type = ? AND word_id IN ?", uid, types.WordTypePhrase, ids).
			Delete(nil).Error; e != nil {
			return errors.ErrorDatabaseDeleteError("删除短语标签关联失败").WithCause(e)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &types.BatchDeleteWordPhraseResp{Deleted: int(deleted)}, nil
}
