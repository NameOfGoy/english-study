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

// maxBatchDeleteWord 单次批量删除单词上限, 防止超大 IN 列表
const maxBatchDeleteWord = 500

type BatchDeleteWordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

// 字典模块/单词批量删除
func NewBatchDeleteWordLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *BatchDeleteWordLogic {
	return &BatchDeleteWordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

// BatchDeleteWord 批量删除当前用户的多个单词及级联数据(词性/学习状态/标签关联), 单事务原子.
// 去重 + 过滤 0; 不存在/非本人的 id 自动忽略; 返回实际删除数. 与 DeleteWord 级联一致, 仅改为 IN.
func (l *BatchDeleteWordLogic) BatchDeleteWord(req *types.BatchDeleteWordReq) (resp *types.BatchDeleteWordResp, err error) {
	ids := dedupIDs(req.IDs)
	if len(ids) == 0 {
		return nil, errors.ErrorRequestParamError("未选择要删除的单词")
	}
	if len(ids) > maxBatchDeleteWord {
		return nil, errors.ErrorRequestParamError("一次最多删除 500 个单词")
	}

	uid := l.ui.ID
	wordTable := fmt.Sprintf("word_user_%d", uid)
	posTable := fmt.Sprintf("word_pos_user_%d", uid)

	var deleted int64
	err = l.svcCtx.Model.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 删用户单词表行; RowsAffected = 实际命中(属于该用户且存在)的数量
		res := tx.Table(wordTable).Where("id IN ?", ids).Delete(nil)
		if res.Error != nil {
			return errors.ErrorDatabaseDeleteError("删除单词失败").WithCause(res.Error)
		}
		deleted = res.RowsAffected
		if deleted == 0 {
			return nil // 没有命中本人单词, 不报错, deleted=0
		}
		// 2. 删词性表 (一对多)
		if e := tx.Table(posTable).Where("word_id IN ?", ids).Delete(nil).Error; e != nil {
			return errors.ErrorDatabaseDeleteError("删除单词词性失败").WithCause(e)
		}
		// 3. 删学习状态 (user_id 隔离)
		if e := tx.Table("word_statuses").
			Where("user_id = ? AND word_type = ? AND word_id IN ?", uid, types.WordTypeWord, ids).
			Delete(nil).Error; e != nil {
			return errors.ErrorDatabaseDeleteError("删除单词学习状态失败").WithCause(e)
		}
		// 4. 删标签关联 (user_id 隔离)
		if e := tx.Table("word_tags").
			Where("user_id = ? AND word_type = ? AND word_id IN ?", uid, types.WordTypeWord, ids).
			Delete(nil).Error; e != nil {
			return errors.ErrorDatabaseDeleteError("删除单词标签关联失败").WithCause(e)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &types.BatchDeleteWordResp{Deleted: int(deleted)}, nil
}

// dedupIDs 去重并过滤 0, 保持顺序. (单词/短语批量删除共用)
func dedupIDs(in []uint) []uint {
	seen := make(map[uint]struct{}, len(in))
	out := make([]uint, 0, len(in))
	for _, id := range in {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
