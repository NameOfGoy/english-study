// Package eventbus 集中声明 english-study 域内的资源事件 DSL.
//
// 设计原则 (对齐 iep / ess-ops-srv 的事件总线规约):
//   - 本包**只**做事件 DSL + 类型化通道定义, **不**放任何业务订阅 handler.
//   - 资源所属模块负责 publish (例如 tag 模块自己发 tag.deleted).
//   - 依赖该资源的模块在自己包内 subscribe 并处理 (例如 wordtagsub 包订阅
//     tag.deleted 后清理 word_tags), 这样依赖关系永远是单向 (subscriber → eventbus),
//     避免 eventbus 反过来 import 业务包导致的循环.
//
// 使用:
//
//	// publish (在资源所属模块里)
//	eventbus.Tag().Deleted().PublishAsync(eventbus.TagDeletedPayload{...})
//
//	// subscribe (在依赖该事件的模块里)
//	eventbus.Tag().Deleted().SubscribeAsync(func(p TagDeletedPayload, _ *generator.EventMetadata) {
//	    // 自己模块的清理 / 联动逻辑
//	}, false)
//
// 修改 EventsDSL 后必须重新生成 events.go + typed.go:
//
//	eventbus-gen -dsl=internal/eventbus/dsl.go
package eventbus

import (
	"github.com/NameOfGoy/resource-eventbus/generator"
)

// TagDeletedPayload 标签被删除事件载荷.
//   - TagID:    被删除的 tag.id
//   - IsSystem: true 表示删除的是系统标签 (tag.is_system=true); 此时 UserID 字段无意义,
//     订阅者应清掉**所有用户**对该 tag 的关联
//   - UserID:   IsSystem=false 时, 该标签的拥有者 user_id; 订阅者据此**只**清这个用户的关联,
//     避免误清同名异主标签
type TagDeletedPayload struct {
	TagID    uint
	IsSystem bool
	UserID   uint
}

// EventsDSL 由 eventbus-gen 静态解析 (不会被运行时调用).
// 返回值必须是 []generator.ResourceDef 字面量, 不能引用变量或调用函数.
func EventsDSL() []generator.ResourceDef {
	return []generator.ResourceDef{
		{
			Resource: "tag",
			Operations: []generator.OpDef{
				{Op: "deleted", Struct: TagDeletedPayload{}},
			},
		},
	}
}
