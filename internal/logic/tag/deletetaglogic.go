package tag

import (
	"context"

	"english-study/internal/errors"
	"english-study/internal/eventbus"
	"english-study/internal/utils"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTagLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewDeleteTagLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *DeleteTagLogic {
	return &DeleteTagLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *DeleteTagLogic) DeleteTag(req *types.DeleteTagReq) (*types.DeleteTagResp, error) {
	tg := l.svcCtx.Model.Gen.Tag

	// 1. 取目标 tag, 校验归属 (系统标签仅超管, 用户标签仅本人)
	target, err := tg.WithContext(l.ctx).Where(tg.ID.Eq(req.ID)).First()
	if err != nil {
		return nil, errors.ErrorDatabaseQueryError("标签不存在").WithCause(err)
	}
	if !canMutateTag(l.ui, target.UserID) {
		return nil, errors.ErrorPermissionError("无权删除该标签")
	}

	// 2. 删 tag (不在此处级联清 word_tags, 那一步交给 eventbus 订阅者异步做; 系统标签会涉及所有用户的关联, 同步做拖长接口)
	if _, err := tg.WithContext(l.ctx).Where(tg.ID.Eq(target.ID)).Delete(); err != nil {
		return nil, errors.ErrorDatabaseDeleteError("删除标签失败").WithCause(err)
	}

	// 3. 发事件: 订阅者跑 DELETE FROM word_tags WHERE tag_id=? [AND user_id=?]
	//    最坏情况(进程崩) 由启动时 ReapOrphanWordTags 兜底
	eventbus.TagDeleted.PublishAsync(eventbus.TagDeletedPayload{
		TagID:    target.ID,
		IsSystem: target.UserID == 0,
		UserID:   target.UserID,
	})

	return &types.DeleteTagResp{}, nil
}
