// word_tags 关联表对外部资源事件的订阅 + 启动维护.
//
// 放在 dictionary 包是因为 word_tags 本身是 dictionary 域的关联表 (单词↔标签),
// 它对 tag 资源变动的响应天然属于 dictionary 这边的收尾工作.
//
// 规约对齐 iep / ess-ops-srv: internal/eventbus 包只放 DSL + 生成的 typed 通道,
// 不写业务订阅 handler; 由依赖该事件的业务模块在自己包内挂载订阅.
//
// 当前订阅:
//   - eventbus.Tag().Deleted() → 同步清掉对应 word_tags 行
//
// 启动钩子:
//   - ReapOrphanWordTags: 兜底删除指向已不存在 tag 的孤儿 word_tags
//     (上一次进程在 PublishAsync 之后挂掉, 订阅者未跑完的场景)
package dictionary

import (
	"context"
	"fmt"

	"english-study/internal/eventbus"

	"github.com/NameOfGoy/resource-eventbus/generator"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// SubscribeTagEvents 把本包的 handler 挂到 eventbus.Tag() 上.
// 用 SubscribeAsync(transactional=false), 多事件并发处理; 失败只 log,
// 启动孤儿清理 (ReapOrphanWordTags) 作为兜底.
func SubscribeTagEvents(db *gorm.DB) error {
	if err := eventbus.Tag().Deleted().SubscribeAsync(makeTagDeletedHandler(db), false); err != nil {
		return fmt.Errorf("subscribe Tag().Deleted(): %w", err)
	}
	return nil
}

// makeTagDeletedHandler 工厂; 暴露内层函数方便单测时绕过 SubscribeAsync 直接调用.
// 用 db.Exec 而非 GORM 链式 Delete, 避免隐式事务包装 (Begin/Commit), 单测期望更直观.
func makeTagDeletedHandler(db *gorm.DB) func(eventbus.TagDeletedPayload, *generator.EventMetadata) {
	return func(p eventbus.TagDeletedPayload, _ *generator.EventMetadata) {
		ctx := context.Background()
		var res *gorm.DB
		if p.IsSystem {
			res = db.WithContext(ctx).Exec(`DELETE FROM "word_tags" WHERE tag_id = ?`, p.TagID)
		} else {
			res = db.WithContext(ctx).Exec(`DELETE FROM "word_tags" WHERE tag_id = ? AND user_id = ?`, p.TagID, p.UserID)
		}
		if res.Error != nil {
			logx.Errorf("dictionary.tagDeleted: 清理 word_tags 失败 tag_id=%d is_system=%v user_id=%d err=%v",
				p.TagID, p.IsSystem, p.UserID, res.Error)
			return
		}
		logx.Infof("dictionary.tagDeleted: 清理完成 tag_id=%d is_system=%v rows=%d",
			p.TagID, p.IsSystem, res.RowsAffected)
	}
}

// ReapOrphanWordTags 启动时孤儿清理: 删 word_tags 中指向已不存在的 tag_id 的行.
// 兜底 PublishAsync 后进程意外死掉导致订阅者没跑完的情况.
func ReapOrphanWordTags(db *gorm.DB) error {
	res := db.Exec("DELETE FROM word_tags WHERE tag_id NOT IN (SELECT id FROM tags)")
	if res.Error != nil {
		return fmt.Errorf("reap orphan word_tags: %w", res.Error)
	}
	if res.RowsAffected > 0 {
		logx.Infof("dictionary: 启动孤儿 word_tags 清理删除 %d 行", res.RowsAffected)
	}
	return nil
}
