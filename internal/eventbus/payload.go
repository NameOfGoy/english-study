// Package eventbus 集中定义 english-study 域内的事件类型 + Typed 事件实例.
// 业务侧 (logic / svc) 只直接 import 这个包, 不再 import resource-eventbus 本体,
// 保留以后换底层 bus 的余地.
package eventbus

import (
	resbus "github.com/NameOfGoy/resource-eventbus"
)

// TagDeletedPayload 标签被删除事件载荷.
//   - TagID:    被删除的 tag.id
//   - IsSystem: true 表示删除的是系统标签 (tag.user_id=0), 此时 UserID 字段无意义
//   - UserID:   IsSystem=false 时, 该标签的拥有者 user_id; 订阅者据此只清这个用户的关联
type TagDeletedPayload struct {
	TagID    uint
	IsSystem bool
	UserID   uint
}

// TagDeleted 标签删除事件实例; 业务代码 import 后直接用:
//
//	eventbus.TagDeleted.PublishAsync(eventbus.TagDeletedPayload{...})
//	eventbus.TagDeleted.SubscribeAsync(myHandler, false)
var TagDeleted = &resbus.TypedEvent[TagDeletedPayload]{Topic: "tag.deleted"}
