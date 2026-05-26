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

type DeleteWordPhraseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewDeleteWordPhraseLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *DeleteWordPhraseLogic {
	return &DeleteWordPhraseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

// DeleteWordPhrase 删除当前用户的某个短语及级联数据.
// 单事务原子: 用户短语表行 + 学习状态行 + 标签关联行.
// (OSS 图片暂不清理, 同 DeleteWord.)
func (l *DeleteWordPhraseLogic) DeleteWordPhrase(req *types.DeleteWordPhraseReq) (resp *types.DeleteWordPhraseResp, err error) {
	if req.ID == 0 {
		return nil, errors.ErrorRequestParamError("短语ID不能为空")
	}

	uid := l.ui.ID
	phraseTable := fmt.Sprintf("word_phrase_user_%d", uid)

	err = l.svcCtx.Model.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		var cnt int64
		if e := tx.Table(phraseTable).Where("id = ?", req.ID).Count(&cnt).Error; e != nil {
			return errors.ErrorDatabaseQueryError("校验短语归属失败").WithCause(e)
		}
		if cnt == 0 {
			return errors.ErrorRequestParamError("短语不存在或已被删除")
		}

		if e := tx.Table(phraseTable).Where("id = ?", req.ID).Delete(nil).Error; e != nil {
			return errors.ErrorDatabaseDeleteError("删除短语失败").WithCause(e)
		}
		if e := tx.Table("word_statuses").
			Where("user_id = ? AND word_type = ? AND word_id = ?", uid, types.WordTypePhrase, req.ID).
			Delete(nil).Error; e != nil {
			return errors.ErrorDatabaseDeleteError("删除短语学习状态失败").WithCause(e)
		}
		if e := tx.Table("word_tags").
			Where("user_id = ? AND word_type = ? AND word_id = ?", uid, types.WordTypePhrase, req.ID).
			Delete(nil).Error; e != nil {
			return errors.ErrorDatabaseDeleteError("删除短语标签关联失败").WithCause(e)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &types.DeleteWordPhraseResp{}, nil
}
