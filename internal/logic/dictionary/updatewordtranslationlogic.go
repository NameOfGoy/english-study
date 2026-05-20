package dictionary

import (
	"context"
	"fmt"
	"strings"

	"english-study/internal/errors"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdateWordTranslationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewUpdateWordTranslationLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *UpdateWordTranslationLogic {
	return &UpdateWordTranslationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *UpdateWordTranslationLogic) UpdateWordTranslation(req *types.UpdateWordTranslationReq) (*types.UpdateWordTranslationResp, error) {
	if len(req.Items) == 0 {
		return nil, errors.ErrorRequestParamError("items 不能为空")
	}

	// 至少一条非空，避免词条全部失去释义
	nonEmpty := 0
	for _, it := range req.Items {
		if strings.TrimSpace(it.Translation) != "" {
			nonEmpty++
		}
		if it.WordPosID == 0 {
			return nil, errors.ErrorRequestParamError("word_pos_id 不能为0")
		}
	}
	if nonEmpty == 0 {
		return nil, errors.ErrorRequestParamError("至少保留一条非空释义")
	}

	// 找到这些 word_pos_id 对应的 word_id，再校验它们都属于当前用户的词典
	posTable := fmt.Sprintf("word_pos_user_%d", l.ui.ID)
	posIDs := make([]uint, 0, len(req.Items))
	for _, it := range req.Items {
		posIDs = append(posIDs, it.WordPosID)
	}

	type posRow struct {
		ID uint
	}
	var found []posRow
	if err := l.svcCtx.Model.DB.WithContext(l.ctx).
		Table(posTable).
		Where("id IN ?", posIDs).
		Find(&found).Error; err != nil {
		return nil, errors.ErrorDatabaseQueryError("校验 pos 归属失败").WithCause(err)
	}
	if len(found) != len(posIDs) {
		return nil, errors.ErrorRequestParamError("部分 word_pos_id 不存在或不属于当前用户")
	}

	// 事务批量更新
	err := l.svcCtx.Model.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		for _, it := range req.Items {
			trans := strings.TrimSpace(it.Translation)
			if err := tx.Table(posTable).
				Where("id = ?", it.WordPosID).
				Update("translation", trans).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, errors.ErrorDatabaseUpdateError("更新释义失败").WithCause(err)
	}

	return &types.UpdateWordTranslationResp{}, nil
}
