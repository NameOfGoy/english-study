package eventbus

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// RegisterSubscribers 在 svc 初始化时注册所有事件订阅者. 当前包括:
//   - TagDeleted → 级联清 word_tags
//
// 订阅采用 SubscribeAsync(transactional=false), 多事件并发处理; 失败只 log,
// 不重试 (按设计 §9, 启动时孤儿清理兜底).
func RegisterSubscribers(db *gorm.DB) error {
	if err := TagDeleted.SubscribeAsync(makeTagDeletedHandler(db), false); err != nil {
		return fmt.Errorf("subscribe TagDeleted: %w", err)
	}
	return nil
}

// makeTagDeletedHandler 工厂; 暴露内层函数方便单测时绕过 SubscribeAsync 直接调用.
// 用 db.Exec 而非 GORM 链式 Delete, 避免隐式事务包装 (Begin/Commit), 单测期望更直观.
func makeTagDeletedHandler(db *gorm.DB) func(TagDeletedPayload) {
	return func(p TagDeletedPayload) {
		ctx := context.Background()
		var res *gorm.DB
		if p.IsSystem {
			res = db.WithContext(ctx).Exec(`DELETE FROM "word_tags" WHERE tag_id = ?`, p.TagID)
		} else {
			res = db.WithContext(ctx).Exec(`DELETE FROM "word_tags" WHERE tag_id = ? AND user_id = ?`, p.TagID, p.UserID)
		}
		if res.Error != nil {
			logx.Errorf("TagDeleted subscriber: 清理 word_tags 失败 tag_id=%d is_system=%v user_id=%d err=%v",
				p.TagID, p.IsSystem, p.UserID, res.Error)
			return
		}
		logx.Infof("TagDeleted subscriber: 清理完成 tag_id=%d is_system=%v rows=%d",
			p.TagID, p.IsSystem, res.RowsAffected)
	}
}

// ReapOrphanWordTags 启动时孤儿清理: 删 word_tags 中指向已不存在的 tag_id 的行.
// 兜底 PublishAsync 后进程意外死掉导致订阅者没跑完的情况 (设计 §4.3).
func ReapOrphanWordTags(db *gorm.DB) error {
	res := db.Exec("DELETE FROM word_tags WHERE tag_id NOT IN (SELECT id FROM tags)")
	if res.Error != nil {
		return fmt.Errorf("reap orphan word_tags: %w", res.Error)
	}
	if res.RowsAffected > 0 {
		logx.Infof("启动时孤儿 word_tags 清理: 删除 %d 行", res.RowsAffected)
	}
	return nil
}
