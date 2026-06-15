# English Study 系统架构文档

## 1. 系统概览

英语学习系统，Go后端(go-zero) + Vue3前端(Vant H5)。

**核心业务:** 单词从导入到掌握的完整生命周期管理，结合AI生成内容(图片、发音、例句)。

## 2. 单词状态流转 (核心业务逻辑)

```
Excel导入 → STUDY(1) →完成学习→ REVIEW(2) →复习通过(≥5轮)→ FINISH(4)
                                    ↓不通过                      ↓抽查不通过
                                STRENGTHEN(3) ←──────────────────┘
                                    ↓通过(≥3轮)
                                  REVIEW(2)
```

**规则细节:**
- Review: 每个单词每天只能复习一次(StudyTime < today)，需连续通过5轮
- Strengthen: 连续"记得"3轮后回到Review重新开始
- Spot: 每次抽查50个，按完成复习时间比例分配(3月+:1月+:2周内 = 30:15:5)

## 3. 后端架构

### 3.1 技术栈
- 框架: go-zero
- 数据库: PostgreSQL + GORM (+ GORM Gen 强类型查询)
- 对象存储: MinIO
- AI: 智谱BigModel(LLM+图片生成) + 阿里云NLS(TTS)
- 认证: JWT (HMAC-SHA256, 含 user_id/username/role 三个 claim) + 微信OAuth
- 密码: bcrypt(cost=12); 兼容历史 PBKDF2 / 明文, 登录成功自动升级
- 进程内事件总线: [github.com/NameOfGoy/resource-eventbus](https://github.com/NameOfGoy/resource-eventbus) (typed Pub/Sub + 可选 DSL 代码生成)

### 3.2 目录结构
```
englishstudy.go              # 入口，初始化所有服务
api/*.api                    # API定义(goctl格式)
internal/
├── handler/                 # 生成的HTTP处理器
├── logic/                   # 业务逻辑(手写)
│   ├── dictionary/          # 词典CRUD、导入导出
│   │   └── word/            # 单词/短语信息操作
│   ├── practise/            # 四种练习模式
│   ├── tag/                 # 标签管理
│   ├── user/                # 用户认证
│   └── fileservice/         # 文件上传下载
├── svc/                     # 服务上下文(DI容器)
├── config/                  # 配置结构体
├── model/                   # GORM模型
│   ├── bean/                # 实体定义
│   └── dto/                 # GORM Gen生成的查询代码
├── dictionary/              # 词典接口+实现
│   └── impl/                # 具体实现(查词、添加词)
├── AI/
│   ├── llm/                 # LLM接口+BigModel实现
│   ├── tts/                 # 阿里云TTS
│   └── view/                # 图片生成
├── aiapplication/           # AI应用层(例句、图片、发音、翻译)
├── eventbus/                # resource-eventbus 集成: 域内 typed 事件 (TagDeleted) + 订阅者
├── handler/middleware/      # SecurityHeaders 中间件 (X-Content-Type-Options 等)
├── oss/                     # MinIO封装 (含 SafeKeyPart 防路径遍历)
└── wx/                      # 微信登录
```

### 3.3 API端点

#### 用户 `/api/v1/user`
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | /login | 账号密码登录 | 否 |
| POST | /register | 注册 | 否 |
| POST | /login/wx | 微信登录 | 否 |
| POST | /register/wx | 微信注册 | 否 |
| GET | /:id | 获取用户信息 | JWT |
| PUT | /:id | 更新用户信息 | JWT |

#### 词典 `/api/v1/dictionary`
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /word/list | 单词列表(分页、前缀筛选) |
| GET | /word/detail | 单词详情(含词性、例句) |
| POST | /word/add | 添加单词(可自动生成图片) |
| POST | /word/update | 更新单词 |
| POST | /word/delete | 删除单词 |
| POST | /word/picture | AI生成单词图片 |
| POST | /word/picture/update | 更新图片链接 |
| POST | /word/picture/delete | 删除图片 |
| POST | /word/example/generate | AI生成例句 |
| POST | /word/example/update | 更新例句 |
| GET | /phrase/list | 短语列表 |
| GET | /phrase/detail | 短语详情 |
| POST | /phrase/add~delete | 短语CRUD |
| POST | /phrase/picture/* | 短语图片操作 |
| POST | /phrase/example/* | 短语例句操作 |
| GET | /status/list | 单词学习状态 |
| POST | /status/update | 更新学习状态 |
| GET | /tag/list | 单词标签 |
| POST | /tag/update | 更新标签 |
| POST | /operation/import | 批量导入(异步, 支持 [tag]/[---]/inline 多标签) |
| POST | /operation/export | 导出 |

#### 标签 `/api/v1/tag` (独立于词典)
| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| POST | /add | 新建; `is_system=true` 仅超管 | JWT |
| POST | /update | 改名/改色; 系统标签仅超管, 用户标签仅本人 | JWT |
| POST | /delete | 删除; 校验后异步通过 eventbus 级联清 word_tags | JWT |
| GET | /list | 列出系统标签 + 当前用户标签 (超管登录时去重) | JWT |
| GET | /detail | 详情; 可见范围 = 自己的 + 系统 | JWT |

#### 练习 `/api/v1/practise`
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /study/list | 获取学习单词卡; `?tag_ids=1,2` 全局标签筛选 |
| POST | /study/finish | 完成学习 |
| GET | /review/list | 获取复习单词卡(仅StudyTime<today); 同样支持 tag_ids |
| POST | /review/finish | 复习完成/失败 |
| GET | /strength/list | 获取加强记忆单词卡; 同样支持 tag_ids |
| POST | /strength/finish | 加强完成/失败 |
| GET | /spot/list | 获取抽查单词卡; 同样支持 tag_ids |
| POST | /spot/finish | 抽查通过/失败 |

#### 分享 `/api/v1/share`
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /generate | 生成 5min HMAC-SHA256 签名 token (可按 tag 筛选) |
| POST | /preview | 解 token + 预览源用户的词条/标签 |
| POST | /word-detail | 拿源用户某词的完整详情 |
| POST | /import | 导入; `tag_import_mode: 0-不带 / 1-仅系统 / 2-全部` |

#### 仪表盘 `/api/v1/dashboard`
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | / | 聚合学习/复习/加强/抽查可用数 + 今日完成 + 总进度 (单次 SQL FILTER) |

### 3.4 数据模型

**全局表:**
- `users` - 用户(ID, Account, WxOpenID, Username, Password, Phone, Email, Avatar, **Role** 0=普通 1=超管)
- `word` - 词典单词(ID, Word, 美音音标, 美音音频, 英音音标, 英音音频, **Source** stardict|ai)
- `word_pos` - 词性(ID, WordID, Word, Pos, Translation, Example[JSON], Picture, Exchange[JSON])
- `word_phrase` - 短语(ID, Phrase, Translation, Pronunciation, Example[JSON], Picture)
- `stardict` - 标准词典(外部数据源)
- `tags` - 标签(ID, Tag, Style, UserID); **UserID=0 视为系统标签**, 全局共享; 复合唯一 (UserID, Tag)

**用户表(动态创建 `{table}_user_{userId}`):**
- `word_user_{uid}` - 用户单词副本
- `word_pos_user_{uid}` - 用户词性副本
- `word_phrase_user_{uid}` - 用户短语副本

**关联表:**
- `word_statuses` - 学习状态(WordID, WordType, Status, Times, Weight, StudyTime, **SRS 字段**(EaseFactor, Interval, NextReviewAt, Repetitions), UserID, **SourceUserID**); 复合唯一 (UserID, WordID, WordType)
- `word_tags` - 单词标签关联(WordID, WordType, TagID, UserID)

**WordType:** 1=单词, 2=短语
**Pos枚举:** 1-n, 2-vt, 3-vi, 4-adv, 5-adj, 6-prep, 7-conj, 8-interj, 9-part, 10-pron, 11-num, 12-art, 13-aux

**迁移脚本** (`docs/migrations/`):
- 001 SRS 字段
- 002 backfill next_review_at
- 003 默认 8 个系统标签 (user_id=0)
- 004 修复 null example 脏数据
- 005 用户表例句从主词典回填
- 006 热路径索引 (pg_trgm GIN, word_statuses 复合, import_tasks 等)
- 007 word_statuses / tags 唯一约束
- 008 users.role 字段 + id=0 升超管
- 009 word_tags(user_id, tag_id) include (word_id, word_type) 给练习全局筛选用

### 3.5 AI服务架构

```
aiapplication/
├── wordexample/     # 例句生成 → BigModel LLM (JSON输出)
├── wordpicture/     # 图片生成 → BigModel Image API → MinIO
├── wordpronounce/   # 发音生成 → 阿里云NLS TTS → MinIO
└── wordtranslation/ # 翻译 → BigModel LLM
```

### 3.6 关键实现

- **单词导入:**
  - 异步 goroutine 处理, 逐行解析后自动查 StarDict、生成语音/例句/图片
  - 文本格式 (见 `docs/import-format.txt`): 支持 `[tag]` 段标签 / `[---]` 终止符 / 行末 `[t1] [t2]` 多内联标签
  - 解析在 `internal/logic/dictionary/import_parser.go` (`ParseImportLines` + `peelInlineTags`), 全 table-driven 单测覆盖
  - 同名标签优先级: 系统 → 当前用户已有 → 自动新建用户级
- **并发控制:** Dictionary 用 RWMutex + adding map 防止重复并发添加同一单词
- **密码:**
  - 新密码用 bcrypt(cost=12), 存储 `$2a$12$...`
  - `VerifyPassword` 兼容 PBKDF2 旧记录 (`<hex>:<salt>`) 与明文; 验证成功且非 bcrypt 时返回 `needUpgrade=true`, 登录 logic 自动 rehash 并 UPDATE
- **MinIO 路径:** `picture/user_word_{uid}/{SafeKeyPart(word)}/{pos}_{ts}.png`, `pronounce/{SafeKeyPart(word)}/{accent}.mp3`; SafeKeyPart 剔除非 `[a-z0-9_-]` 字符防越权
- **标签权限:** 系统标签 (`user_id=0`) 仅超管可改; 用户标签仅本人 (`canMutateTag` helper). 服务端鉴权一律读 JWT 解出的 role, 不信任请求体的任何 role 字段
- **标签删除一致性:** `DeleteTagLogic` → 提交事务删 `tags` 行 → `eventbus.Tag().Deleted().PublishAsync(...)` 链式发布 → 订阅者 (`internal/logic/dictionary/wordtag_subscriber.go`, 与 word_tag 其他 logic 同包) 异步 `DELETE FROM word_tags WHERE tag_id=? [AND user_id=?]`; 启动时 main 函数调用 `dictlogic.ReapOrphanWordTags` 兜底 (清掉指向已删 tag 的孤儿关联). `internal/eventbus` 只放 DSL + 生成的 typed 通道, 不写业务订阅 handler — 订阅器与该数据所属业务包同处, 挂载点在 main 而非 model 包 (避免 model→logic→svc→model 循环)
- **分享导入:** HMAC-SHA256 全 32 字节签名 + uint64 nonce, 5 分钟 TTL; 导入模式 0/1/2 分别对应 不带 / 仅系统 / 全部 (用户级走字符串匹配, 缺则新建; 系统级 tag_id 全局共享, 直接复用)
- **练习全局筛选:** 4 个 list 接 `tag_ids` 后 → 反向 SELECT word_tags 拿 (word_id, word_type) → IN 限制主查询; 索引 `(user_id, tag_id) INCLUDE (word_id, word_type)`
- **资源 cascade 删除:** DeleteWord/DeleteWordPhrase 单事务清: 用户表行 + 词性行 + word_statuses + word_tags
- **HTTP 安全头:** 全局 middleware 加 `X-Content-Type-Options: nosniff` / `X-Frame-Options: DENY` / `Referrer-Policy: no-referrer`
- **proxyImage SSRF 防御:** 白名单 + 自定义 `http.Transport.DialContext` 在 dial 时 reresolve IP 并拒绝私网/loopback, 闭合 DNS rebinding 时间窗; io.Copy 限到 20MB

## 4. 前端架构

### 4.1 技术栈
Vue 3.3.4 + Vite + Vue Router + Axios + Vant 4.6.2 + SCSS

### 4.2 目录结构
```
English-Study-UI/src/
├── api/          # user.js, dictionary.js, practise.js, tag.js, file.js, dashboard.js, share.js
├── components/
│   ├── TabBar.vue, SearchModal.vue
│   ├── dictionary/ # WordDetailView/PhraseDetailView/ShareImportModal/TranslationEditModal/...
│   └── practice/   # TagFilterBar.vue, ImageCarousel, ExampleCarousel, ...
├── composables/  # useAudioPlayer, useCardPictureEditor, usePracticeCards, ...
├── router/       # 路由配置 (含 isTokenExpired 主动检查的 beforeEach)
├── utils/        # auth.js (localStorage + role + JWT fallback isAdmin), request.js, practiceTagFilter.js
├── views/        # Dashboard, Login, Register, Profile, Practice(hub+4模式), Dictionary, TagManage, ImportTaskHistory
├── styles/
└── main.js
```

### 4.3 路由
- `/` → `/dashboard` (默认首页, 当日学习/复习/加强/抽查统计 + 总进度)
- `/login`, `/register` (公开)
- `/practice` (练习入口; 顶部 TagFilterBar 全局标签筛选, 跨 4 模式生效)
- `/practice/study|review|strength|spot`
- `/dictionary` (词典)
- `/profile`, `/profile/tags` (标签管理), `/profile/import-tasks` (导入历史)

### 4.4 页面交互

**Dashboard:** 单接口拉聚合统计 + 4 模式卡片 + 总进度条
**Practice (hub):** 4 模式卡片 + 顶部 `TagFilterBar` (折叠/多选+"全部"互斥, localStorage 持久化跨页生效)
**PracticeStudy(学习):** 图片轮播 + 单词 + 英美发音 + 翻译 + 例句轮播 → "完成学习"; 卡片支持图片编辑(生成/上传/搜图/裁剪)、释义即时编辑
**PracticeReview(复习):** 3 阶段显示(图片→提示→完整) → "已掌握"/"未掌握"; SRS SM-2 推算下次复习时间
**PracticeStrength(加强):** 显示中文 → "记得"/"不记得" → 进度条; 连续 3 次"记得"后回 Review
**PracticeSpot(抽查):** 类似复习流程
**Dictionary:** 3 种视图(字母/状态/标签) + 搜索 + 词/短语切换; 单词详情页右上角红色垃圾桶 = 删除单词
**TagManage:** 系统标签 + 用户标签分组展示; 超管对系统标签也能编辑/删除, 普通用户系统标签段没操作按钮
**ShareImportModal:** 输入分享码 → 预览 + 3 种标签导入模式 radio (不带/仅系统/全部) → 后台异步导入

### 4.5 权限与 role
- 登录响应 `data.role` (0=普通, 1=超管) 写入 localStorage 后供 UI 显隐
- `utils/auth.js` `isAdmin()` 先读 localStorage, 缺失时回退解码 JWT claims 并回填 (兼容老 session)
- **真正的权限校验全部在后端**, 前端 role 仅用于按钮/段落显隐

### 4.5 部署
Docker + Nginx, dev代理 `/api` → `localhost:8888`

## 5. 前后端联调

- 前端 baseURL: `/api/v1`
- 认证: JWT Bearer token (含 user_id/username/role 三个 claim)
- 401 → 自动跳登录页
- 资源 URL: 走当前页面 origin, 由 nginx 或 Vite proxy 转发到后端/MinIO
- nginx: SPA fallback (`location /`) 显式 `Cache-Control: no-cache`, 避免新部署后旧 index.html 被浏览器缓存导致 JS 分包 404

## 6. 事件总线 (resource-eventbus)

**位置:** [github.com/NameOfGoy/resource-eventbus](https://github.com/NameOfGoy/resource-eventbus) @ v0.2.0

**特点:**
- 轻量 in-process 总线, 封装 `asaskevich/EventBus`
- 提供 `TypedEvent[T]` 泛型, 既能手写也能通过 `eventbus-gen` CLI 从 DSL 生成 (含 metadata: source/trace_id/timestamp)

**包职责划分 (对齐 iep / ess-ops-srv 的规约):**
- `internal/eventbus/dsl.go` — 用 `EventsDSL() []generator.ResourceDef` 声明全部资源 + 操作 + payload struct
- `internal/eventbus/events.go` + `typed.go` — 由 `eventbus-gen -dsl=internal/eventbus/dsl.go` 生成, 提供 `eventbus.Tag().Deleted()` 链式 DSL, **不要手改**
- 业务订阅 handler **不**放在 eventbus 包里, 由依赖该资源的业务模块**与该数据所属同包**实现 (例如 `internal/logic/dictionary/wordtag_subscriber.go` 与 getwordtaglistlogic 等同处 dictionary 包). 依赖方向永远 subscriber → eventbus, 避免反向 import
- **挂载点在 main**: 应用启动时 `englishstudy.go` 显式调用 `dictlogic.ReapOrphanWordTags(m.DB)` + `dictlogic.SubscribeTagEvents(m.DB)`. 不在 model 包内挂, 因为 model 被 svc 依赖, svc 被 logic 依赖, 若 model 再 import logic 就形成循环

**当前事件:**
- `Tag().Deleted() → TagDeletedPayload{TagID, IsSystem, UserID}` — DeleteTagLogic publish, dictionary 包同 logic 文件订阅级联清 `word_tags`

**故障语义:** `PublishAsync` 后进程崩 → 关联未清; main 启动调用 `dictlogic.ReapOrphanWordTags` 兜底 `DELETE FROM word_tags WHERE tag_id NOT IN (SELECT id FROM tags)`
