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

type DeleteWordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

// 字典模块/单词删除
func NewDeleteWordLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *DeleteWordLogic {
	return &DeleteWordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

// DeleteWord 删除当前用户的某个单词及其级联数据.
// 单事务原子: 用户单词表行 + 词性表行 + 学习状态行 + 标签关联行.
// (OSS 上的图片暂不清理, 视为孤儿; 以后做个定时清理任务再扫.)
func (l *DeleteWordLogic) DeleteWord(req *types.DeleteWordReq) (resp *types.DeleteWordResp, err error) {
	if req.ID == 0 {
		return nil, errors.ErrorRequestParamError("单词ID不能为空")
	}

	uid := l.ui.ID
	wordTable := fmt.Sprintf("word_user_%d", uid)
	posTable := fmt.Sprintf("word_pos_user_%d", uid)

	err = l.svcCtx.Model.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 校验单词归属 (查 user 表里这条记录是否存在; 不存在=不属于该用户/已删, 直接拒)
		var cnt int64
		if e := tx.Table(wordTable).Where("id = ?", req.ID).Count(&cnt).Error; e != nil {
			return errors.ErrorDatabaseQueryError("校验单词归属失败").WithCause(e)
		}
		if cnt == 0 {
			return errors.ErrorRequestParamError("单词不存在或已被删除")
		}

		// 2. 删用户单词表行
		if e := tx.Table(wordTable).Where("id = ?", req.ID).Delete(nil).Error; e != nil {
			return errors.ErrorDatabaseDeleteError("删除单词失败").WithCause(e)
		}
		// 3. 删用户词性表所有相关行 (一对多)
		if e := tx.Table(posTable).Where("word_id = ?", req.ID).Delete(nil).Error; e != nil {
			return errors.ErrorDatabaseDeleteError("删除单词词性失败").WithCause(e)
		}
		// 4. 删 word_statuses (该用户 + 单词类型 + word_id)
		if e := tx.Table("word_statuses").
			Where("user_id = ? AND word_type = ? AND word_id = ?", uid, types.WordTypeWord, req.ID).
			Delete(nil).Error; e != nil {
			return errors.ErrorDatabaseDeleteError("删除单词学习状态失败").WithCause(e)
		}
		// 5. 删 word_tags 关联
		if e := tx.Table("word_tags").
			Where("user_id = ? AND word_type = ? AND word_id = ?", uid, types.WordTypeWord, req.ID).
			Delete(nil).Error; e != nil {
			return errors.ErrorDatabaseDeleteError("删除单词标签关联失败").WithCause(e)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &types.DeleteWordResp{}, nil
}
