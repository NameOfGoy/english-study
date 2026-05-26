# 标签系统改造设计

> 状态：**草案 v0.1**，等用户审阅。审阅通过后按"实现顺序"分批落地。
> 涉及：后端 (go-zero) + 前端 (Vue3) + 新仓库 `resource-eventbus` + DB 迁移。

---

## 0. 目标速览

1. **标签分级**：系统级 (超管控) + 用户级 (用户自管)。普通用户对系统标签**只能用、不能改**。
2. **分享导入分模式**：全部 / 仅系统 / 不带。系统标签直接 ID 引用；用户标签按字符串匹配，缺则新建。
3. **标签删除一致性**：用内部消息总线广播 `tag.deleted` 事件，触发 `word_tags` 级联清理。系统级删除影响所有用户行。
4. **导入文本支持标签段**：`[tag]` 开始打标、`[---]` 结束、`[other]` 切换。
5. **练习页全局标签筛选**：多选 + 全部，对四个模式生效。

---

## 1. 数据模型变更

### 1.1 `bean.User` — 新增 `Role`

```go
type User struct {
    // ... 既有字段
    Role int `gorm:"column:role;not null;default:0;index;comment:角色 0-普通 1-超管"`
}
```

迁移：`docs/migrations/008_user_role.sql`

```sql
ALTER TABLE users ADD COLUMN IF NOT EXISTS role INT NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
-- 把现有 user_id=1 升为超管 (单用户项目, 第一个注册的就是你)
UPDATE users SET role = 1 WHERE id = 1;
```

JWT claim 添加 `role`，前端可读但仅作 UI 显隐，**权限校验必须在后端**。

### 1.2 `bean.Tag` — 字段不动，语义沿用

- 仍然 `user_id = 0` 表示系统标签（与 `003_default_tags.sql` 兼容）
- 不引入 `is_system` 列，避免双写不一致
- 已有 unique index `idx_tag_user_name (user_id, tag)` 继续生效

### 1.3 `bean.WordTag` — 字段不动

- 仍是单表 `word_tags`，按 `user_id` 隔离用户的"贴标签"动作
- 系统标签的 `WordTag` 行：`tag_id` 指向 `user_id=0` 的 tag，但 `word_tags.user_id` 仍是贴标签的用户自己

---

## 2. Tag CRUD 权限

| 操作 | 普通用户 | 超管 |
|---|---|---|
| 列出系统标签 | ✅ | ✅ |
| 列出自己的标签 | ✅ | ✅ |
| 列出他人的标签 | ❌ | ❌（前端没入口，不主动开放） |
| 新建用户标签 (自己) | ✅ | ✅ |
| 新建系统标签 | ❌ | ✅ |
| 编辑用户标签 (自己) | ✅ | ✅ |
| 编辑系统标签 | ❌ | ✅ |
| 删除用户标签 (自己) | ✅ | ✅ |
| 删除系统标签 | ❌ | ✅ |
| 给词条贴系统/用户标签 | ✅ | ✅ |

实现要点：
- 现有 `addtaglogic` / `updatetaglogic` / `deletetaglogic` 加入 `userIsAdmin` 守卫
- 区分 `target_user_id`：
  - 操作系统标签 (`Tag.UserID == 0`)：要求 `caller.Role == 1`
  - 操作用户标签：要求 `tag.UserID == caller.ID`
- 中间件层提供 `utils.RequireAdmin(ctx)` 工具方法

---

## 3. resource-eventbus 库

**仓库**：`github.com/NameOfGoy/resource-eventbus`（公开，MIT）
**Module**：`github.com/NameOfGoy/resource-eventbus`
**作用域**：纯通用 in-process 事件总线抽象，**不含任何业务**。

### 3.1 包结构

```
resource-eventbus/
├── README.md
├── LICENSE (MIT)
├── go.mod
├── eventbus.go        // Bus 接口 + 默认实现 (基于 sync.Map + chan)
├── singleton.go       // GetBus() 单例 + SetBus() 测试注入
├── typed.go           // TypedEvent[T] 泛型类型安全包装
└── example_test.go    // 用例
```

### 3.2 `Bus` 接口

```go
type Bus interface {
    Publish(topic string, args ...any)
    Subscribe(topic string, fn any) error
    SubscribeAsync(topic string, fn any, transactional bool) error
    SubscribeOnce(topic string, fn any) error
    Unsubscribe(topic string, handler any) error
    HasCallback(topic string) bool
    WaitAsync()
}
```

实现：直接套 `github.com/asaskevich/EventBus`（成熟、零依赖、MIT），不重造轮子。

### 3.3 `TypedEvent[T]` 泛型包装

```go
type TypedEvent[T any] struct {
    Topic string
}

func (e *TypedEvent[T]) Publish(data T)               { GetBus().Publish(e.Topic, data) }
func (e *TypedEvent[T]) PublishAsync(data T)          { go e.Publish(data) }
func (e *TypedEvent[T]) Subscribe(handler func(T)) error
func (e *TypedEvent[T]) SubscribeAsync(handler func(T), transactional bool) error
```

不抄 ess-ops-srv 的 codegen DSL，**保持库轻量**；业务方自己声明 `TypedEvent` 实例即可。

### 3.4 业务侧（english-study）使用方式

```
internal/eventbus/
├── bus.go        // var TagDeleted = &resourceeventbus.TypedEvent[TagDeletedPayload]{Topic:"tag.deleted"}
└── payload.go    // type TagDeletedPayload struct { TagID uint; IsSystem bool; UserID uint }
```

`go.mod` 加 `require github.com/NameOfGoy/resource-eventbus v0.1.0`。

---

## 4. Tag 删除一致性流程

### 4.1 事件定义

```go
type TagDeletedPayload struct {
    TagID    uint
    IsSystem bool  // tag.user_id == 0
    UserID   uint  // 仅 IsSystem=false 时有效
}

var TagDeleted = &resourceeventbus.TypedEvent[TagDeletedPayload]{Topic: "tag.deleted"}
```

### 4.2 删除流程

**主流程 (`DeleteTagLogic.DeleteTag`)：**

1. 鉴权（权限矩阵见 §2）
2. 起事务：
   - `DELETE FROM tags WHERE id = ?`
3. 事务提交成功后，`TagDeleted.PublishAsync(...)`
4. 同步返回 ok（关联清理由订阅者后台异步做）

**订阅者 (`internal/eventbus/subscriber/tag_subscriber.go`)：**

```go
TagDeleted.SubscribeAsync(func(p TagDeletedPayload) {
    if p.IsSystem {
        // 系统标签: 清掉所有用户的 word_tags 关联
        db.Exec("DELETE FROM word_tags WHERE tag_id = ?", p.TagID)
    } else {
        // 用户标签: 只清这个用户的
        db.Exec("DELETE FROM word_tags WHERE tag_id = ? AND user_id = ?", p.TagID, p.UserID)
    }
    logx.Infof("tag.deleted reaped, tag_id=%d is_system=%v rows=...", p.TagID, p.IsSystem)
}, false)
```

订阅者在 `svc.NewServiceContext` 初始化时注册一次。

### 4.3 故障语义

- **删除事务成功，事件未消费**（进程在 Publish 后秒死）：
  - 重启时 `svc` 重新注册订阅，但已 Publish 的事件丢失
  - 副作用：`tags` 表无该行，`word_tags` 残留指向不存在的 tag_id
  - **兜底**：启动时跑一次 `DELETE FROM word_tags WHERE tag_id NOT IN (SELECT id FROM tags)` 孤儿清理 (放在 `internal/model/db.go` 的启动钩子)
- **删除事务失败**：tags 行还在，事件未发布，无副作用
- **订阅者执行失败**：log error，下次启动孤儿清理兜底

> 之所以选 in-process eventbus 而非外部 MQ：本项目单进程部署、单用户量级，避免运维负担。

---

## 5. Share 导入 3 模式

### 5.1 API 变更

`api/share.api`：

```diff
 type ImportShareReq {
     Token      string `json:"token"`
-    ImportTags bool   `json:"import_tags,optional"` // 是否同步导入标签
+    TagImportMode int `json:"tag_import_mode,optional"` // 0-不带 1-仅系统 2-全部 (默认 0)
 }
```

### 5.2 后端处理（`importTags`）

```
mode 0 → 跳过 (现状)
mode 1 → 只复制源用户 word_tags 里 tag_id 是系统级 (tags.user_id=0) 的关联
         系统 tag 在所有用户视角下 ID 一致, 直接复用 tag_id
mode 2 → 系统级照 mode 1 处理 + 用户级走字符串匹配 (已实现)
```

### 5.3 前端（`ShareImportModal.vue`）

把现在的 `导入标签` checkbox 改成三选：
```
( ) 不带标签
(•) 仅系统标签
( ) 全部标签 (含对方自建标签)
```

---

## 6. 导入文本格式扩展

### 6.1 文本约定

```
apple
banana
[重点词汇]            ← 开启 "重点词汇" 标签
cat                   ← 打上 "重点词汇"
dog                   ← 打上 "重点词汇"
[考试词汇]            ← 切换到 "考试词汇" (隐式结束上一个)
elephant              ← 打上 "考试词汇"
[---]                 ← 显式结束, 接下来不打标签
frog                  ← 无标签
[新标签x]             ← 隐式新建用户标签 "新标签x"
grape                 ← 打上 "新标签x"
```

规则：
- 任何 `^\[(.+?)\]$` 行视作标签段标记
- 段标记内容 `---` 表示清空当前标签
- 其它内容为标签名，若不存在则自动创建为**当前用户的标签**（不会自动创建系统标签）
- 嵌套不允许（一次只有一个"当前标签"）

### 6.2 实现位置

`internal/logic/dictionary/importwordlogic.go` 的 scanner 循环重构：

```go
type importItem struct {
    Word    string
    TagName string // "" 表示无标签
}
items := parseImportLines(lines)  // 处理 [tag]/[---] 标记, 返回 (word, activeTag) 列表

// 每个 item 处理时:
//   1. 走原有 AddWord/AddPhrase 流程
//   2. 若 TagName != "": 找/建 tag → 写 word_tags
```

新增工具：`internal/logic/dictionary/import_parser.go` + 单元测试。

---

## 7. 练习页全局标签筛选

### 7.1 前端

- `src/views/Practice.vue` (入口页) 顶部加 `<TagFilterBar>` 组件
  - `van-checkbox-group` 多选 + 一个 "全部" 单独项（互斥逻辑）
  - 选中态写 `localStorage('practice_tag_filter', number[])`，跨页保留
- 进入 Study/Review/Strengthen/Spot 任一模式时，把当前 `tag_ids` 作 query param 带给后端

### 7.2 后端

四个 list 接口加同名 query 参数：

```
GET /api/v1/practise/study/word-card/list?tag_ids=1,2,3
GET /api/v1/practise/review/word-card/list?tag_ids=...
... 同样加给 strengthen / spot
```

SQL 改造（`getstudywordcardlistlogic.go` 等四份）：

```sql
SELECT ws.* FROM word_statuses ws
WHERE ws.user_id = ? AND ws.status = ?
  AND (
    $tag_ids IS EMPTY  OR
    EXISTS (
      SELECT 1 FROM word_tags wt
      WHERE wt.user_id = ws.user_id
        AND wt.word_id = ws.word_id
        AND wt.word_type = ws.word_type
        AND wt.tag_id IN (...)
    )
  )
```

为这条增量查询加索引：

```sql
-- migration 009
CREATE INDEX IF NOT EXISTS idx_word_tags_user_word_tag
    ON word_tags (user_id, word_id, word_type, tag_id);
```

---

## 8. 实现顺序 & 迁移

| 批次 | 内容 | 后端 | 前端 | DB 迁移 | 风险 |
|---|---|---|---|---|---|
| **B1** | User.Role + tag CRUD 权限 | bean.User、middleware、tag 三 logic | 标签管理页隐藏编辑/删除按钮 (非超管+系统标签) | 008 | 低 |
| **B2** | resource-eventbus 公开仓库 + 接入 | 新仓库 + go get + svc 装配订阅者 | 无 | 无 | 中 (新仓库) |
| **B3** | Tag 删除 eventbus 流程 + 启动孤儿清理 | DeleteTag + subscriber + db.go reconcile | 无 | 无 | 低 |
| **B4** | Share 导入 3 模式 | importsharelogic 改 mode | ShareImportModal 三选 | 无 | 低 |
| **B5** | 导入文本 [tag]/[---] | import_parser + importwordlogic | 导入页提示文案 | 无 | 低 |
| **B6** | 练习全局标签筛选 | 四个 list logic + 加索引 | TagFilterBar + Practice.vue | 009 | 中 |

每批次单独 commit + 推送，便于 review 与回滚。

---

## 9. 确认结果（2026-05-23）

1. **超管候选**：DB 里已有 `id=0` 用户行，迁移 008 直接 `UPDATE users SET role = 1 WHERE id = 0;`。`id=1` 仍是日常用户。系统标签 `tags.user_id=0` 与之天然同归属。
2. **系统标签预设**：保留 `003_default_tags.sql` 里 8 个默认标签。
3. **导入终止符**：**严格匹配字面量 `[---]`**（大小写敏感，破折号数量必须是 3 个）。其他 `^\[(.+?)\]$` 全部按标签名处理。
4. **eventbus 失败处理**：不加重试；启动时 `internal/model/db.go` 跑一次 `DELETE FROM word_tags WHERE tag_id NOT IN (SELECT id FROM tags)` 兜底。
5. **练习筛选互斥**：选"全部"则清空具体标签数组；勾选任何具体标签则自动取消"全部"。前端在 `TagFilterBar` 组件内部维护这个状态机。
