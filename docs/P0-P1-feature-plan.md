# P0~P1 功能方案总汇

> 5角色 Brainstorm (产品经理 / UI设计师 / 交互设计师 / 前端开发 / 架构师+后端)
> 
> 生成日期: 2026-04-02

---

## 目录

- [总览：实施顺序](#总览实施顺序)
- [功能1: 今日学习看板 (P0)](#功能1-今日学习看板-p0)
- [功能2: 搜索添加单词 (P0)](#功能2-搜索添加单词-p0)
- [功能3: 自适应间隔重复 SRS (P0)](#功能3-自适应间隔重复-srs-p0)
- [功能4: 听音拼写 (P1)](#功能4-听音拼写-p1)
- [功能5: 中译英模式 (P1)](#功能5-中译英模式-p1)
- [功能6: 学习统计+打卡 (P1)](#功能6-学习统计打卡-p1)
- [全局技术决策汇总](#全局技术决策汇总)

---

## 总览：实施顺序

```
Sprint 1 ─── 功能3: SRS自适应间隔重复 (改表结构，后续功能依赖)
Sprint 2 ─── 功能2: 搜索添加单词 (完全独立，去掉Excel门槛)
Sprint 3 ─── 功能1: 今日学习看板 (依赖SRS的next_review_at)
Sprint 4 ─── 功能4+5: 听音拼写 + 中译英 (可并行，结构对称)
Sprint 5 ─── 功能6: 学习统计+打卡 (需在所有finish接口埋点)
```

### 功能依赖关系

```
                    ┌──────────────────────────┐
                    │   功能2: 搜索添加单词      │  ← 完全独立，可最先开发
                    │ (替代Excel导入的入口)       │
                    └──────────┬───────────────┘
                               │ 用户有了词之后
                               ▼
┌─────────────────────────────────────────────────────────┐
│                 功能1: 今日学习看板                        │
│   需要功能3的 next_review_at 来准确统计"今日待复习"        │
└────────────────────────┬────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────┐
│            功能3: 自适应间隔重复 (SRS)                     │
│     修改 WordStatus 表 + FinishReview 状态机               │
└─────────────────────────────────────────────────────────┘
                         │
          ┌──────────────┼──────────────┐
          ▼              ▼              ▼
┌────────────────┐ ┌────────────────┐ ┌─────────────────────┐
│ 功能4: 听音拼写 │ │ 功能5: 中译英   │ │ 功能6: 学习统计+打卡  │
│  独立练习模式    │ │  独立练习模式   │ │  依赖所有练习模式     │
│  需要已有音频    │ │  需要已有释义   │ │  在各finish接口埋点   │
└────────────────┘ └────────────────┘ └─────────────────────┘
```

---

## 功能1: 今日学习看板 (P0)

### 1.1 产品定义

**用户故事**: As a user, I want to 打开应用时立刻看到今天需要做什么（有多少词要学、多少词要复习、多少词要加强），so that 我不需要自己去各个模块里翻找，直接点击就能开始。

**核心场景**: 用户早上打开H5应用，首页显示:
- "今日待学习：12个新词" — status=1(Study)
- "今日待复习：8个" — status=2(Review) 且 next_review_at <= now
- "今日待加强：3个" — status=3(Strengthen) 且 next_review_at <= now
- "已掌握：156个" — status=4(Finish)

用户点击任一卡片，直接跳转到对应模式开始学习。

| MVP做 | MVP不做 |
|-------|---------|
| 聚合接口返回4种状态计数 + 今日已完成数 | 每日目标设置 |
| 前端首页卡片展示 + 点击跳转 | 时间线图表 |
| 新用户全0的兜底UI | 推送提醒 |

**验收标准**:
- [ ] 接口响应 < 200ms（单表聚合查询，有user_id索引）
- [ ] 返回数据与各模块列表接口数据一致
- [ ] 新注册用户（无任何词）看到全部为0，不报错
- [ ] 完成操作后刷新看板数字立刻更新
- [ ] 全部完成时进度环变绿+庆祝动效

### 1.2 UI设计

#### 页面布局

**顶部问候区（高度约 180px）：**
- 背景: `linear-gradient(135deg, #1989fa 0%, #1565c0 100%)` 与现有 practice-header 一致
- 左上角: 问候语"上午好/下午好/晚上好"，font-size 14px，opacity 0.9，白色
- 左侧第二行: 用户昵称，font-size 22px，font-weight 700，白色
- 右上角: 用户头像 48x48 圆形，`van-image`，边框 2px solid rgba(255,255,255,0.5)
- 下方居中: 白色数据卡片"浮出"覆盖蓝色头部底部约40px

**今日数据卡片（悬浮）：**
- 白色卡片，border-radius 16px，box-shadow `0 8px 24px rgba(0,0,0,0.08)`
- 三列数据，flex justify-content: space-around
  - "待学习" — 数字 font-size 28px, font-weight 800, color `#1989fa`
  - "待复习" — 数字同样式, color `#ff9800`
  - "已掌握" — 数字同样式, color `#07c160`
- 列间 1px solid `#ebedf0` 竖分隔线

**今日进度条区域：**
- 标签: "今日进度 3/15"，右侧百分比 "20%"
- `van-progress` 高度 8px，border-radius 4px，颜色 `#1989fa`
- 进度条下方小文字说明"再学2个即可达到今日目标"

**快速开始区域：**
- 标题"快速开始"，font-size 16px, font-weight 600
- 水平滚动卡片: `overflow-x: auto; scroll-snap-type: x mandatory`
- 每张卡片 140x100px, border-radius 12px, margin-right 12px
- 四张卡片: 学习`#e3f2fd` / 复习`#fff3e0` / 强化`#fbe9e7` / 抽查`#ede7f6`
- 左侧4px竖色条，对应模式主色

**最近学习记录：**
- `van-cell-group` 最近3-5个学过的词
- title=英文单词，label=中文释义，value=学习时间("10分钟前")

#### 状态变化

| 状态 | 视觉表现 |
|------|---------|
| 加载中 | `van-skeleton` 骨架屏（3列方块+圆形+长条） |
| 空数据(新用户) | 数字全0，进度条0%，引导按钮"去添加单词" |
| 全部完成 | 进度条100%变绿色`#07c160`，文字"太棒了，今日目标已完成！" |

#### Vant组件
- `van-image`（头像）
- `van-progress`（今日进度）
- `van-cell` / `van-cell-group`（最近学习列表）
- `van-skeleton`（加载态骨架屏）
- `van-empty`（空态）

### 1.3 交互设计

#### 动效

| 动效 | 描述 |
|------|------|
| 进度环填充 | 0% → 实际百分比，800ms `ease-out`，SVG `stroke-dashoffset` |
| 数字跳动 | 0递增到目标值，600ms，`requestAnimationFrame` |
| 角标弹跳 | 模式卡片角标 `scale(0)→scale(1)`，弹跳曲线，依次延迟200ms |
| 满环庆祝 | 蓝→绿过渡(500ms) + 12个半透明粒子扩散消失 |
| 数据变化 | 数字短暂放大 scale(1.2) 再恢复 |

#### 操作流程

| 手势 | 区域 | 行为 |
|------|------|------|
| 点击 | 模式卡片 | 进入对应练习模式 |
| 左滑 | 整个页面 | 切换到词典页（沿用现有逻辑） |
| 下拉 | 整个页面 | `van-pull-refresh` 刷新看板数据 |

- 从子模式返回时自动静默刷新（`onActivated` + `<keep-alive>`）
- 网络错误: 看板显示"数据加载失败"+重试按钮，模式卡片仍可点击

### 1.4 前端方案

#### 文件清单

| 操作 | 文件路径 |
|------|----------|
| 新增 | `src/api/dashboard.js` |
| 新增 | `src/views/Dashboard.vue` |
| 新增 | `src/components/dashboard/TodayTaskCard.vue` |
| 新增 | `src/components/dashboard/StudyProgressRing.vue` (纯SVG) |
| 新增 | `src/components/dashboard/ModeEntryGrid.vue` |
| 修改 | `src/router/index.js` — `/` 重定向到 `/dashboard` |
| 修改 | `src/components/TabBar.vue` — 增加"首页"tab |

#### API调用

```js
// src/api/dashboard.js
// GET /v1/dashboard/today
// 返回: { study: {total, finished}, review: {total, finished},
//         strength: {total, finished}, spot: {total, finished},
//         total_words, streak_days }
export function getTodayDashboard() {
  return request({ url: '/v1/dashboard/today', method: 'get' })
}
```

#### 路由变更

```js
{ path: '/', redirect: '/dashboard' },
{
  path: '/dashboard',
  name: 'Dashboard',
  component: () => import('@/views/Dashboard.vue'),
  meta: { title: '首页', requiresAuth: true, showTabbar: true }
}
```

TabBar从三个tab变为四个: **首页** / 练习 / 词典 / 我的

#### 关键实现

- 环形进度: SVG `<circle>` + `stroke-dasharray/stroke-dashoffset`，不引入图表库
- 缓存: `sessionStorage` key=`dashboard_today_{date}` 做1分钟TTL短缓存
- 不需要引入任何新依赖

### 1.5 后端方案

#### API设计

```
GET /api/v1/dashboard

Response:
{
  "data": {
    "study_count": 12,
    "review_count": 8,
    "strengthen_count": 3,
    "spot_count": 156,
    "today_studied": 5,
    "today_reviewed": 3,
    "today_strengthened": 1,
    "total_words": 320,
    "finished_words": 156,
    "progress_rate": 0.4875
  }
}
```

#### 核心逻辑

```go
// 单条SQL用PostgreSQL FILTER语法一次聚合所有count
var result AggResult
err := db.Model(&bean.WordStatus{}).
    Select(`
        COUNT(*) AS total_words,
        COUNT(*) FILTER (WHERE status = 1) AS study_count,
        COUNT(*) FILTER (WHERE status = 2 AND next_review_at <= ?) AS review_count,
        COUNT(*) FILTER (WHERE status = 3 AND next_review_at <= ?) AS strengthen_count,
        COUNT(*) FILTER (WHERE status = 4) AS spot_count,
        COUNT(*) FILTER (WHERE study_time >= ? AND study_time < ?) AS today_studied,
        ...
    `, now, now, today, tomorrow).
    Where("user_id = ?", uid).
    Scan(&result).Error
```

#### 性能考虑

- PostgreSQL `FILTER (WHERE ...)` 单次全表扫描，比7次独立COUNT快
- 建议索引: `CREATE INDEX idx_ws_user_status ON word_status(user_id, status, next_review_at)`
- 高频接口（每次进首页），考虑内存缓存 TTL 30秒

#### 影响文件

| 文件 | 变更 |
|------|------|
| `api/englishstudy.api` | 添加 `import "dashboard.api"` |
| `api/dashboard.api` | 新增文件 |
| `internal/logic/dashboard/getdashboardlogic.go` | 手动编写业务逻辑 |
| handler/routes/types | `make api` 自动生成 |

无新表、无迁移。

---

## 功能2: 搜索添加单词 (P0)

### 2.1 产品定义

**用户故事**: As a user, I want to 在应用内搜索一个英文单词并一键添加到我的词库，so that 我不需要准备Excel文件、上传、等待导入任务完成。

**核心场景**:
1. 点击"添加单词"，进入搜索页
2. 输入 "eph"，实时显示 stardict 中匹配的词(ephemeral, ephemeron...)，展示单词+中文释义+音标
3. 点击"ephemeral"可查看完整信息
4. 点击"添加到词库"，后端走 `Dictionary.AddWord()` 流程，单词进入"学习"状态
5. 添加成功后按钮变灰"已添加"，可继续搜索下一个

| MVP做 | MVP不做 |
|-------|---------|
| stardict前缀搜索(limit 20) | 模糊纠错(输入ephmeral不建议ephemeral) |
| 一键添加复用 Dictionary.AddWord() | 批量选择添加 |
| 标记已添加状态 | stardict中不存在的词处理 |
| 支持中文搜索translation字段 | — |

**验收标准**:
- [ ] 搜索结果 500ms 内返回（stardict.sw有索引）
- [ ] 搜索结果展示单词、中文释义、音标
- [ ] 已添加的词标记"已添加"，不可重复添加
- [ ] 空输入不发请求，< 2个字符不搜索
- [ ] `Dictionary.IsWordAdding()` 返回true时显示"添加中"

### 2.2 UI设计

#### 入口

- Dictionary.vue 头部右侧圆形按钮 36x36，`rgba(255,255,255,0.25)`
- `van-icon name="plus"` 白色 size 20

#### 搜索页面（全屏）

**顶部搜索栏（固定）：**
- 白色背景 56px，底部 1px solid `#ebedf0`
- 返回箭头 + `van-search` shape="round" placeholder="输入英文单词搜索词库"
- 聚焦时底部 2px `#1989fa` 线

**热门区域（搜索框为空时）：**
- "热门单词"标题 + `van-tag` 标签云(flex-wrap, gap 8px)

**搜索结果列表：**
- `van-list` 无限滚动，每条 72px
- 左侧: 单词(16px粗) + 音标(12px灰)
- 中间: 中文释义(13px灰，1行截断)
- 右侧: 未添加→蓝色圆形+按钮(36x36) / 已添加→灰色tag"已添加"

**添加动效：**
- +按钮→`van-loading`旋转→绿色勾号(scale弹跳200ms)→"已添加"tag
- 底部 `showToast({ message: '已添加到词典', type: 'success' })`

#### 状态变化

| 状态 | 视觉表现 |
|------|---------|
| 初始(刚进入) | 搜索框自动聚焦弹键盘，显示热门标签云 |
| 搜索中(debounce后) | 3-5个 `van-skeleton` 行骨架 |
| 无结果 | `van-empty` description="词库中未找到相关单词" |
| 网络错误 | `van-empty` image="network" description="网络异常，请重试" |

### 2.3 交互设计

#### 操作流程

```
词典页点击+号 → 全屏Popup底部滑入(300ms)
→ 搜索框自动聚焦 → 输入关键词(debounce 300ms)
→ 显示搜索结果 → 点击+号添加
→ 按钮变loading → 变绿色勾号 → 变"已添加"
→ 可继续搜索添加更多 → 返回后词典页自动刷新
```

| 手势 | 区域 | 行为 |
|------|------|------|
| 点击 | 返回箭头/"取消" | 关闭面板 |
| 点击 | 结果行(非按钮区) | 展开详细释义(accordion) |
| 点击 | "+"按钮 | 触发添加 |
| 上下滑动 | 结果列表 | 滚动浏览 |
| 下拉 | 面板顶部 | 关闭面板 |

#### 反馈机制
- 添加成功: 行内状态变化(不用Toast遮挡)，整行短暂左移4px弹回
- 添加失败: 按钮恢复为+，行下方红色小字"添加失败，点击重试"，3秒消失
- 已添加提示: 点击"已添加"行，Toast"该单词已在你的词典中"
- 键盘收起: 点击列表区域自动 input.blur()

### 2.4 前端方案

#### 文件清单

| 操作 | 文件路径 |
|------|----------|
| 新增 | `src/api/stardict.js` |
| 新增 | `src/views/SearchWord.vue` |
| 新增 | `src/components/search/SearchResultCard.vue` |
| 新增 | `src/components/search/StardictWordDetail.vue` |
| 修改 | `src/router/index.js` — 新增 `/search` 路由 |
| 修改 | `src/views/Dictionary.vue` — 头部增加搜索入口按钮 |

#### API调用

```js
// src/api/stardict.js

// 搜索stardict词库
// GET /v1/dictionary/search/word
export function searchStardict(params) {
  return request({ url: '/v1/dictionary/search/word', method: 'get', params })
}

// 从stardict一键添加到个人词库
// POST /v1/dictionary/search/add
export function addStardictToLibrary(data) {
  return request({ url: '/v1/dictionary/search/add', method: 'post', data })
}
```

#### 路由

```js
{
  path: '/search',
  name: 'SearchWord',
  component: () => import('@/views/SearchWord.vue'),
  meta: { title: '搜索单词', requiresAuth: true, showTabbar: false }
}
```

#### 关键实现

- 防抖: 手写 debounce 300ms，避免引入 lodash
- 输入法兼容: 监听 `compositionstart/compositionend` 设 `isComposing` flag
- 滚动加载: `van-list` offset累加分页
- 添加成功后本地更新 `already_added`，不重新搜索
- 不需要引入任何新依赖

### 2.5 后端方案

#### API设计

```
GET /api/v1/dictionary/search/word?keyword=xxx&limit=20

Response:
{
  "data": [
    {
      "word": "abandon",
      "phonetic": "/əˈbændən/",
      "translation": "vt. 放弃；遗弃\nn. 放任",
      "tag": "cet4 cet6 ielts",
      "collins": 5,
      "is_added": false
    }
  ]
}

POST /api/v1/dictionary/search/add
Body: { "word": "abandon" }
→ 复用 word.WordInfo.GetCustomizedWordInfo + IncreaseWord
```

#### 核心逻辑

```go
func (l *SearchWordLogic) SearchWord(req *types.SearchWordReq) (*types.SearchWordResp, error) {
    keyword := strings.TrimSpace(strings.ToLower(req.Keyword))

    // 判断中英文
    if isChineseKeyword(keyword) {
        query = query.Where(sdg.Translation.Like("%" + keyword + "%"))
    } else {
        query = query.Where(sdg.Sw.Like(keyword + "%")) // 前缀匹配
    }

    results, _ := query.Order(sdg.Frq.Desc()).Limit(limit).Find()

    // 批量查 word_user_{userId} 标记 is_added
    addedMap := l.getAddedWordsMap(wordList)

    // ... 组装返回
}
```

#### 性能考虑

- stardict表约40万+词条，前缀搜索用 `btree varchar_pattern_ops` 索引可达毫秒级
- `getAddedWordsMap` 用 `IN` 查 word_user 表，批量一次解决
- AI生成(TTS/例句/图片)在 AddWord 中同步完成(3-8秒)，前端显示loading

#### 需要的索引

```sql
CREATE INDEX IF NOT EXISTS idx_stardict_sw_prefix
ON stardict USING btree (sw varchar_pattern_ops);
```

#### 影响文件

| 文件 | 变更 |
|------|------|
| `api/dictionary.api` | 追加 search group 路由和类型 |
| `internal/logic/dictionary/searchwordlogic.go` | 新增 |
| `internal/logic/dictionary/quickaddwordlogic.go` | 新增 |
| 现有 AddWordLogic | **不修改**，复用 word.WordInfo |

---

## 功能3: 自适应间隔重复 SRS (P0)

### 3.1 产品定义

**用户故事**: As a user, I want to 系统根据我每次复习的表现自动调整下次复习时间，so that 我记得牢的词不用天天复习浪费时间，记不住的词能更频繁地出现。

**核心场景改进**:

| 对比 | 原系统 | SRS改进后 |
|------|--------|----------|
| 复习轮次 | 固定5轮 | 动态，由SM-2算法决定 |
| 复习间隔 | 每天1次 | 1天→6天→17天→48天... 指数增长 |
| 掌握标准 | 复习5次 | 间隔>=30天 且 连续正确>=4次 |
| 失败处理 | 直接进强化 | 间隔重置为1天+进强化 |

| MVP做 | MVP不做 |
|-------|---------|
| SM-2算法替代固定5轮 | 完整Anki参数调优 |
| word_status加4个字段 | "加强"模式SRS化 |
| 迁移现有数据 | 提前复习功能 |
| 前端展示下次复习时间 | 难度系数个性化调整 |

**验收标准**:
- [ ] 复习成功后，该词在下次复习到期前不出现在列表中
- [ ] 复习失败后，该词明天再次出现
- [ ] 间隔序列大致为: 1天→6天→17天→48天→128天
- [ ] 间隔>=30天且连续正确>=4次自动转为FINISH
- [ ] 现有数据迁移后不会突然看到所有词都需复习

### 3.2 SM-2 算法

#### 数学公式

```
如果回答正确(quality >= 3):
  repetitions++
  if repetitions == 1: interval = 1天
  if repetitions == 2: interval = 6天
  if repetitions >= 3: interval = ceil(interval × EF)

如果回答错误(quality < 3):
  repetitions = 0
  interval = 1天

EF' = EF + (0.1 - (5 - quality) × (0.08 + (5 - quality) × 0.02))
EF' = max(EF', 1.3)    // 下限1.3
interval = min(interval, 180)  // 上限180天
```

#### Quality评分映射

| 分数 | 含义 | 前端操作 |
|------|------|---------|
| 5 | 完美记忆，毫不犹豫 | "记得"(默认) |
| 4 | 正确但有犹豫 | "模糊" |
| 3 | 困难但正确 | "困难" |
| 2 | 错误 | "忘了"/"不记得" |
| 1 | 几乎忘了 | — |
| 0 | 完全空白 | — |

#### 间隔增长示例 (全部perfect, quality=5)

| 次数 | EF | 间隔 | 下次复习 |
|------|-----|------|---------|
| 1 | 2.60 | 1天 | 明天 |
| 2 | 2.70 | 6天 | 一周后 |
| 3 | 2.80 | 17天 | 两周半后 |
| 4 | 2.80 | 48天 | 一个半月后 |
| 5 | 2.80 | 128天 | → FINISH |

#### Go实现

```go
// internal/logic/practise/srs.go

func SM2Calculate(easeFactor float64, interval int, repetitions int, quality int) SRSResult {
    q := float64(quality)
    newEF := easeFactor + (0.1 - (5-q)*(0.08+(5-q)*0.02))
    if newEF < 1.3 {
        newEF = 1.3
    }

    var newInterval int
    var newRepetitions int

    if quality >= 3 {
        newRepetitions = repetitions + 1
        switch newRepetitions {
        case 1:
            newInterval = 1
        case 2:
            newInterval = 6
        default:
            newInterval = int(math.Ceil(float64(interval) * newEF))
        }
    } else {
        newRepetitions = 0
        newInterval = 1
    }

    if newInterval > 180 {
        newInterval = 180
    }

    nextReview := time.Now().Truncate(24*time.Hour).
        Add(time.Duration(newInterval) * 24 * time.Hour)

    return SRSResult{
        EaseFactor:  math.Round(newEF*100) / 100,
        Interval:    newInterval,
        Repetitions: newRepetitions,
        NextReview:  nextReview,
    }
}
```

### 3.3 数据模型变更

#### word_status 新增字段

```go
// internal/model/bean/word_status.go 新增字段

EaseFactor   float64   `gorm:"column:ease_factor;default:2.5;comment:难度因子(SM-2)"`
Interval     int       `gorm:"column:interval;default:0;comment:当前复习间隔(天)"`
NextReviewAt time.Time `gorm:"column:next_review_at;index;comment:下次复习时间"`
Repetitions  int       `gorm:"column:repetitions;default:0;comment:连续正确次数"`
```

#### DDL

```sql
ALTER TABLE word_status
    ADD COLUMN IF NOT EXISTS ease_factor DOUBLE PRECISION NOT NULL DEFAULT 2.5,
    ADD COLUMN IF NOT EXISTS interval INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS next_review_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS repetitions INTEGER NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_word_status_next_review
    ON word_status(user_id, status, next_review_at);
```

### 3.4 状态机改造

```
原: review成功 → times++ → times>=5 → FINISH
新: review成功 → SM2Calculate → interval>=30 && reps>=4 → FINISH
    review失败 → STRENGTHEN(interval重置为1)
    strengthen成功 → 回REVIEW(EF保留，interval=1重新开始)
```

#### API变更

```
FinishReviewReq 新增字段:
  quality int `json:"quality,optional,default=4"` // SM-2质量评分 3-5

Review列表查询条件变更:
  原: study_time < today
  新: next_review_at <= now()
  排序: next_review_at ASC (最紧急的排前面)
```

### 3.5 UI展示

**复习卡片信息条（full阶段）：**
- 高度 36px, background `#fff8e1`, border-radius 8px
- `van-icon name="clock-o"` + "间隔 3天 | 第2次复习"
- 逾期时: background `#ffebee`, color `#ee0a24`, "已逾期1天 | 建议立即复习"

**词典状态视图圆点：**
- 绿色: 距下次复习 > 3天（记忆稳固）
- 黄色: 1-3天内需复习
- 红色: 已过期未复习（脉冲动画每2秒闪一下）

**Toast改进：**
- "已掌握"→"已掌握! 下次复习: 3天后"
- "不记得"→"已标记，稍后将再次出现"

### 3.6 数据迁移

```sql
-- 根据已有times推算SRS参数
UPDATE word_status SET
    ease_factor = 2.5,
    repetitions = times,
    interval = CASE
        WHEN times = 0 THEN 1
        WHEN times = 1 THEN 1
        WHEN times = 2 THEN 6
        ELSE LEAST(POWER(2.5, times - 2) * 6, 180)::int
    END,
    next_review_at = CASE
        WHEN study_time IS NOT NULL THEN study_time + INTERVAL '1 day' * CASE
            WHEN times <= 1 THEN 1
            WHEN times = 2 THEN 6
            ELSE LEAST(POWER(2.5, times - 2) * 6, 180)::int
        END
        ELSE NOW()
    END
WHERE status IN (2, 3); -- REVIEW 和 STRENGTHEN

UPDATE word_status SET next_review_at = NOW() + INTERVAL '90 days'
WHERE status = 4; -- FINISH
```

### 3.7 影响文件

| 文件 | 变更 |
|------|------|
| `internal/model/bean/word_status.go` | 添加4个SRS字段 |
| `internal/logic/practise/srs.go` | **新增** SM-2算法 |
| `internal/logic/practise/srs_test.go` | **新增** 单元测试 |
| `internal/logic/practise/finishreviewlogic.go` | 集成SRS计算 |
| `internal/logic/practise/finishstrengthlogic.go` | 强化通过时设SRS初始参数 |
| `internal/logic/practise/finishstudylogic.go` | Study完成后初始化SRS参数 |
| `internal/logic/practise/finishspotlogic.go` | 抽查失败时重置SRS |
| `internal/logic/practise/getreviewwordcardlistlogic.go` | 查询条件改用next_review_at |
| `internal/logic/practise/getstrengthwordcardlistlogic.go` | 同上 |
| `api/practise.api` | FinishReviewReq增加Quality字段 |
| `internal/model/migration.go` | **新增** 迁移脚本 |

---

## 功能4: 听音拼写 (P1)

### 4.1 产品定义

**用户故事**: As a user, I want to 听到单词的发音后尝试拼写出来，so that 我能训练听力辨识和拼写能力，而不只是看着英文选中文。

**核心场景**:
1. 进入"听音拼写"，系统从 Review/Finish 词中抽取10个
2. 屏幕只显示播放按钮和拼写格子，不显示英文
3. 用户听发音后输入拼写
4. 拼错时高亮差异字母，显示正确答案
5. 完成后显示本轮得分(8/10)和错误词汇列表

| MVP做 | MVP不做 |
|-------|---------|
| 播放TTS+格子拼写+后端校验 | 语音输入 |
| 返回Levenshtein相似度 | 自动加入加强列表 |
| 首字母提示(可选) | 短语听写 |
| 无音频的词自动跳过 | 实时逐字母提示 |

**独立性**: 不影响SRS状态流转，纯额外练习。

**验收标准**:
- [ ] 播放按钮点击后音频1秒内开始播放
- [ ] 大小写不影响判定
- [ ] 一轮结束展示统计（正确数/总数+错误词列表）
- [ ] 无音频文件的单词自动跳过
- [ ] 每个词可重复播放发音（不限次数）

### 4.2 UI设计

#### 配色: 青色系

```scss
.practice-dictation {
  --mode-color: #00bcd4;
  --mode-color-light: #e0f7fa;
  --mode-gradient: linear-gradient(135deg, #e0f7fa 0%, #f5fdff 50%, #ffffff 100%);
}
```

#### 页面布局

**音频播放区域：**
- 大号播放按钮 80x80 圆形，gradient `#00bcd4→#0097a7`
- 播放时: 内部图标切换为音波动画(3条竖线高度变化CSS animation)
- 外圈涟漪扩散 `box-shadow: 0 0 0 12px rgba(0,188,212,0.15)`
- 下方两个小按钮: "慢速" + "再听一次"

**拼写格子区域：**
- 每格 40x48px, border-radius 8px, border 2px solid `#e8e8e8`, gap 4px
- 格数 = 字母数，居中排列，超8个字母格子缩小为 32x40
- 当前焦点格: border-color `#00bcd4`, background `rgba(0,188,212,0.05)`
- 已填入字母: font-size 22px, font-weight 700, text-align center
- 使用隐藏 `<input>` 接收键盘输入，JS映射到格子

#### 状态变化

| 状态 | 视觉表现 |
|------|---------|
| 输入中 | 当前格闪烁光标(blink animation)，已输入格border变青色 |
| 正确 | 所有格子变绿色bg + 光泽从左到右扫过(200ms) + confetti |
| 错误 | 错误格子变红抖动(shake 200ms×2)，显示正确答案 |
| 跳过 | 格子逐个填入正确字母(打字机效果50ms间隔)，灰色 |

### 4.3 交互设计

#### 操作流程

```
进入 → 自动播放发音 → 用户在格子中拼写
→ 填满后出现"确认"按钮 → 提交
→ 正确: 绿色+1.5s后自动下一词
→ 错误: 红色抖动+显示正确答案 → "跳过"/"重试"
```

| 手势 | 区域 | 行为 |
|------|------|------|
| 点击 | 播放按钮 | 播放/重播发音 |
| 点击 | 已填格子 | 将光标移到该格子(允许修改) |
| 键盘 | Backspace | 删除光标前一个字母，光标左移 |
| 点击 | "提示"按钮 | 揭示一个随机字母(最多2次) |

#### 动效

| 动效 | 描述 |
|------|------|
| 播放按钮涟漪 | 3圈 `scale(1)→scale(1.5)` + `opacity 1→0`，循环 |
| 字母填入 | `scale(0) opacity(0)` → `scale(1.1)→scale(1) opacity(1)`，150ms |
| 正确判定 | 全部格子同时变绿，光泽扫过 |
| 错误判定 | `translateX(-4px)→translateX(4px)`，200ms重复2次 |
| 键盘按键 | 按下 `scale(0.92)` 50ms，bg短暂变深 |

### 4.4 前端方案

#### 文件清单

| 操作 | 文件路径 |
|------|----------|
| 新增 | `src/views/PracticeListening.vue` |
| 新增 | `src/components/practice/SpellingInput.vue` |
| 新增 | `src/components/practice/SpellingResult.vue` |
| 新增 | `src/composables/useSpellingCheck.js` |
| 修改 | `src/router/index.js` — 新增 `/practice/listening` |
| 修改 | `src/views/Practice.vue` — mode-grid增加"听写"入口 |
| 修改 | `src/api/practise.js` — 增加听写接口 |

#### 拼写校验(前端)

```js
// src/composables/useSpellingCheck.js
export function useSpellingCheck() {
  const check = (input, correct) => {
    const a = input.trim().toLowerCase()
    const b = correct.trim().toLowerCase()
    const isCorrect = a === b
    const maxLen = Math.max(a.length, b.length)
    const diff = Array.from({ length: maxLen }, (_, i) => ({
      input: a[i] || '',
      correct: b[i] || '',
      match: (a[i] || '').toLowerCase() === (b[i] || '').toLowerCase()
    }))
    return { isCorrect, diff }
  }
  return { check }
}
```

#### 关键实现

- 音频自动播放: 移动端限制，第一张卡需用户点"开始"触发，后续可自动播放
- 虚拟键盘遮挡: `window.visualViewport.resize` 事件动态调整 padding-bottom
- 复用: `PracticeHeader`, `LoaderOverlay`, `CompletionOverlay`, `usePracticeCards`, `useAudioPlayer`

### 4.5 后端方案

#### API设计

```
GET /api/v1/practise/listen-spell/list?count=10&source=0
→ 从Review+Finish状态随机抽取，过滤audio非空
→ 返回:
{
  "data": [
    {
      "word_id": 123,
      "word_type": 1,
      "audio_url": "http://.../pronounce/ephemeral/us.mp3",
      "hint_length": 9,
      "hint_first_char": "e",
      "translation": "adj. 短暂的"
    }
  ]
}

POST /api/v1/practise/listen-spell/submit
Body: { "word_id": 123, "word_type": 1, "spelling": "ephemral" }
→ 后端 toLowerCase+trim 比对
→ 返回:
{
  "data": {
    "correct": false,
    "correct_word": "ephemeral",
    "user_spelling": "ephemral",
    "similarity": 0.89
  }
}
```

#### 核心逻辑

- Levenshtein距离算法: O(m*n)，单词长度通常<30，可忽略
- 相似度 = 1 - distance/max(len(a), len(b))
- 音频URL直接用MinIO链接，无需后端中转

#### 影响文件

| 文件 | 变更 |
|------|------|
| `api/practise.api` | 追加 listen-spell group |
| `internal/logic/practise/getlistenspelllistlogic.go` | 新增 |
| `internal/logic/practise/submitlistenspelllogic.go` | 新增 |
| `internal/logic/practise/spelling.go` | 新增(通用拼写校验+Levenshtein) |

---

## 功能5: 中译英模式 (P1)

### 5.1 产品定义

**用户故事**: As a user, I want to 看到中文释义后回忆并输入对应的英文单词，so that 我能训练从"含义"到"单词"的反向记忆通路。

**核心场景**:
1. 屏幕显示"adj. 短暂的，转瞬即逝的"
2. 用户输入 "ephemeral"，正确，显示音标+例句
3. 如不会，可请求提示: 第一次→首尾字母"e______l"，第二次→完整答案

| MVP做 | MVP不做 |
|-------|---------|
| 显示中文+词性，用户输入英文 | 同义词/近义词判定 |
| 后端精确匹配校验 | 选择题模式 |
| 两级提示(首尾字母→完整) | 短语中译英 |
| 配图作为额外线索(可选) | — |

**验收标准**:
- [ ] 中文释义包含词性标记
- [ ] 多个词性释义时只展示一个(避免答案太明显)
- [ ] 正确后展示完整单词信息(音标+例句+图片)
- [ ] "提示"分两级: 首尾字母→完整答案
- [ ] 一轮结束展示统计

### 5.2 UI设计

#### 配色: 紫色系

```scss
.practice-translate {
  --mode-color: #ab47bc;
  --mode-color-light: #f3e5f5;
  --mode-gradient: linear-gradient(135deg, #f3e5f5 0%, #faf5fc 50%, #ffffff 100%);
}
```

#### 三阶段流程

**阶段一 - 回忆：**
- 卡片顶部: 中文释义 font-size 18px, font-weight 600 (核心展示)
- 词性标签: `van-tag` type="primary" plain round (点"提示"后才显示)
- 底部按钮: "想起来了" | "想不起来"

**阶段二 - 验证（点"想起来了"）：**
- 输入框 `van-field` 48px高，border-radius 12px，聚焦时border变紫
- 字体 font-size 18px, font-weight 600, text-align center

**阶段三 - 完整展示：**
- 复用学习模式卡片: ImageCarousel + PhoneticDisplay + 释义 + ExampleCarousel
- 底部: "已掌握" | "需加强"

#### 状态变化

| 状态 | 视觉表现 |
|------|---------|
| 正确 | 输入框border变绿+bg `#e8f5e9`，单词卡从下滑入(300ms)+自动播放发音 |
| 接近(编辑距离<=2) | 输入框border变黄，"接近了！正确拼写是xxx" |
| 错误 | 输入框border变红+抖动，对比展示(用户答案删除线+正确答案粗体绿) |
| 使用提示后答对 | 同正确，但加小tag"使用了提示" |

### 5.3 交互设计

#### 操作流程

```
显示中文释义 → "想起来了"(弹出输入框) / "想不起来"(直接展示答案)
→ 输入英文 → "确认"
→ 正确/接近/错误反馈 → 展示完整单词卡
→ "已掌握"/"需加强" → 下一词
```

#### 动效

| 动效 | 描述 |
|------|------|
| 中文释义入场 | 打字机效果(每字50ms，最长1秒) |
| 阶段一→二 | 释义上移(200ms)，输入框从下淡入(200ms) |
| 判定正确 | 输入框变绿+上方绿色对号scale弹跳(300ms) |
| 判定错误 | 输入框变红+抖动 |
| 阶段二→三 | `opacity 0 + scale(0.95)` → `opacity 1 + scale(1)` (300ms) |

#### 与现有交互一致性
- 完整展示阶段100%复用学习模式组件组合
- 复用 `usePracticeCards`, `useAudioPlayer`, `useParsedExamples`
- "已掌握"/"需加强" 沿用 `.action-btn.success/.danger` 样式

### 5.4 前端方案

#### 文件清单

| 操作 | 文件路径 |
|------|----------|
| 新增 | `src/views/PracticeTranslate.vue` |
| 修改 | `src/router/index.js` — 新增 `/practice/translate` |
| 修改 | `src/views/Practice.vue` — mode-grid增加"中译英"入口 |
| 修改 | `src/api/practise.js` — 增加中译英接口 |

结构与 `PracticeReview.vue` 极为相似，复制后修改 initial 阶段逻辑即可。

Practice.vue 的 mode-grid 从 2x2 变为 2x3。

### 5.5 后端方案

#### API设计

```
GET /api/v1/practise/cn-to-en/list?count=10
→ 从 status>=2 的词中取
→ 返回:
{
  "data": [
    {
      "word_id": 123,
      "word_type": 1,
      "translation": "adj. 短暂的",
      "pos": "adj.",
      "hint_length": 9,
      "hint_initials": "e",
      "picture": ["http://..."]
    }
  ]
}

POST /api/v1/practise/cn-to-en/submit
Body: { "word_id": 123, "word_type": 1, "answer": "ephemeral" }
→ 复用 CheckSpelling() 通用函数
→ 返回:
{
  "data": {
    "correct": true,
    "correct_word": "ephemeral",
    "similarity": 1.0,
    "audio_url": "http://...",
    "uk_phonetic": "/ɪˈfem.ər.əl/"
  }
}
```

#### 代码复用

听音拼写和中译英共享:

```go
// internal/logic/practise/spelling.go
func CheckSpelling(userInput, correctAnswer string) (correct bool, similarity float64) {
    userInput = strings.TrimSpace(strings.ToLower(userInput))
    correctAnswer = strings.ToLower(correctAnswer)
    correct = userInput == correctAnswer
    similarity = levenshteinSimilarity(userInput, correctAnswer)
    return
}
```

---

## 功能6: 学习统计+打卡 (P1)

### 6.1 产品定义

**用户故事**: As a user, I want to 看到自己每天学了多少词、连续打卡了多少天、各状态的词量变化趋势，so that 我能获得成就感和学习动力。

**核心场景**:
- **打卡日历**: 每个有学习行为的日期标绿点，连续打卡天数高亮
- **今日数据**: 今日新学5个 | 复习8个 | 加强2个
- **累计数据**: 总词库320个 | 已掌握156个 | 学习天数45天
- **打卡规则**: 当天有任意finish/submit操作即自动打卡，无需手动

| MVP做 | MVP不做 |
|-------|---------|
| learning_log明细 + daily_stats日聚合 | 学习时长统计(需前端埋点) |
| 打卡日历热力图 | 折线趋势图(V2) |
| 连续天数 + 累计数据 | 打卡排行榜/社交分享 |
| 所有finish/submit接口埋点 | 补签功能 |

**验收标准**:
- [ ] 每次完成操作时 learning_log 写入一条记录
- [ ] 今日统计数据实时刷新
- [ ] 打卡日历正确标记有学习行为的日期
- [ ] 连续打卡天数计算正确（中间断一天从断后重算）
- [ ] 新用户全0/空日历不报错

### 6.2 数据模型

#### 新增两张表

```go
// study_record - 学习行为明细(append-only)
type StudyRecord struct {
    ID        uint      `gorm:"primaryKey"`
    UserID    uint      `gorm:"column:user_id;index;not null"`
    WordID    uint      `gorm:"column:word_id;not null"`
    WordType  int       `gorm:"column:word_type;not null"` // 1-单词 2-短语
    Mode      int       `gorm:"column:mode;not null"`      // 1-学习 2-复习 3-强化 4-抽查 5-听写 6-中译英
    Result    int       `gorm:"column:result;not null"`    // 1-正确/完成 2-错误/失败
    Quality   int       `gorm:"column:quality;default:0"`  // SM-2评分
    Duration  int       `gorm:"column:duration;default:0"` // 耗时(秒)
    CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;index"`
}

// daily_stats - 每日聚合(UPSERT)
type DailyStats struct {
    ID                uint      `gorm:"primaryKey"`
    UserID            uint      `gorm:"column:user_id;index;not null"`
    Date              time.Time `gorm:"column:date;type:date;not null"`
    StudyCount        int       `gorm:"column:study_count;default:0"`
    ReviewCount       int       `gorm:"column:review_count;default:0"`
    StrengthenCount   int       `gorm:"column:strengthen_count;default:0"`
    SpotCount         int       `gorm:"column:spot_count;default:0"`
    ListenSpellCount  int       `gorm:"column:listen_spell_count;default:0"`
    CnToEnCount       int       `gorm:"column:cn_to_en_count;default:0"`
    TotalCount        int       `gorm:"column:total_count;default:0"`
    CorrectCount      int       `gorm:"column:correct_count;default:0"`
    TotalDuration     int       `gorm:"column:total_duration;default:0"`
    NewWordsCount     int       `gorm:"column:new_words_count;default:0"`
    CreatedAt         time.Time `gorm:"autoCreateTime"`
    UpdatedAt         time.Time `gorm:"autoUpdateTime"`
}
// UNIQUE(user_id, date)
```

#### DDL

```sql
CREATE TABLE IF NOT EXISTS study_record (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    word_id INTEGER NOT NULL,
    word_type INTEGER NOT NULL,
    mode INTEGER NOT NULL,
    result INTEGER NOT NULL,
    quality INTEGER DEFAULT 0,
    duration INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_study_record_user_time ON study_record(user_id, created_at);

CREATE TABLE IF NOT EXISTS daily_stats (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    date DATE NOT NULL,
    study_count INTEGER DEFAULT 0,
    review_count INTEGER DEFAULT 0,
    strengthen_count INTEGER DEFAULT 0,
    spot_count INTEGER DEFAULT 0,
    listen_spell_count INTEGER DEFAULT 0,
    cn_to_en_count INTEGER DEFAULT 0,
    total_count INTEGER DEFAULT 0,
    correct_count INTEGER DEFAULT 0,
    total_duration INTEGER DEFAULT 0,
    new_words_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, date)
);
CREATE INDEX idx_daily_stats_user_date ON daily_stats(user_id, date);
```

### 6.3 UI设计

#### Profile页入口

- 白色圆角卡片 padding 16px, border-radius 12px
- 左侧: 小型热力图预览(7x4色块矩阵，约100x60px)
- 右侧: "学习统计" + "连续打卡7天" + 右箭头

#### 打卡日历热力图 (GitHub Contribution风格)

- 7列(周日-周六) x 5行，每格36x36, border-radius 6px, gap 4px
- 色阶(基于当天学习量):
  - 无记录: `#f0f0f0`
  - 1-5个词: `#c8e6c9`
  - 6-15个词: `#66bb6a`
  - 16-30个词: `#2e7d32`
  - 30+个词: `#1b5e20`
- 今天的格子加 2px solid `#1989fa` 边框
- 未来日期: `#fafafa` + 1px dashed `#e8e8e8`

#### 连续打卡展示条

- background `linear-gradient(135deg, #fff8e1, #fff3e0)`, border-radius 12px
- 火焰emoji + "连续打卡 7 天" + "最长记录: 23天"

#### 数据卡片 (2x2网格)

| 指标 | 颜色 |
|------|------|
| 学习单词数 | `#1989fa` |
| 复习次数 | `#ff9800` |
| 掌握率 | `#07c160` |
| 平均每日 | `#ab47bc` |

#### 状态分布条 (纯CSS)

- 水平堆叠条形图，高度24px, border-radius 12px
- 待学习`#e0e0e0` / 学习中`#1989fa` / 复习中`#ff9800` / 已掌握`#07c160`

### 6.4 交互设计

#### 操作流程

| 手势 | 区域 | 行为 |
|------|------|------|
| 左右滑动 | 日历区域 | 切换月份(CSS transform 300ms) |
| 点击 | 日期格子 | 展开/收起当天学习详情(accordion) |
| 点击 | 统计小卡片 | 弹出趋势详情Sheet |

#### 动效

| 动效 | 描述 |
|------|------|
| 日历月份切换 | `transition: transform 0.3s ease`，滑动翻页感 |
| 绿色圆点出现 | 从月初到月末依次出现(每个20ms)，最多600ms |
| 统计数字 | 计数器动画0→目标值，800ms `ease-out` |
| 打卡庆祝弹窗 | `scale(0.8)+opacity(0)` → `scale(1)+opacity(1)`，300ms + confetti |

#### 打卡庆祝

- 今日首次完成任意练习返回时触发(localStorage记录lastCheckinDate)
- 弹窗居中: "打卡成功！" + "连续打卡 X 天" + 随机鼓励语 + confetti
- 仅弹一次

### 6.5 前端方案

#### 文件清单

| 操作 | 文件路径 |
|------|----------|
| 新增 | `src/api/stats.js` |
| 新增 | `src/views/Statistics.vue` |
| 新增 | `src/components/stats/WeeklyChart.vue` (纯CSS柱状图) |
| 新增 | `src/components/stats/CheckinCalendar.vue` (纯CSS日历) |
| 新增 | `src/components/stats/StatsSummaryCard.vue` |
| 修改 | `src/router/index.js` — 新增 `/statistics` |
| 修改 | `src/views/Profile.vue` — 增加"学习统计"入口cell |

#### API调用

```js
// src/api/stats.js

export function getStatsSummary() {
  return request({ url: '/v1/stats/summary', method: 'get' })
}

export function getDailyStats(params) {
  return request({ url: '/v1/stats/daily', method: 'get', params })
}
```

#### 关键实现

- 柱状图: 纯CSS flexbox + 动态height百分比，不引入ECharts
- 日历: 手写6x7网格，`computed` 计算当月所有日期格子
- 不需要引入任何新依赖

### 6.6 后端方案

#### 埋点方案

在所有 Finish/Submit 方法末尾异步写入:

```go
// internal/logic/practise/record.go

func RecordStudyAction(ctx context.Context, db *gorm.DB, record *bean.StudyRecord) {
    // 1. 插入明细记录
    db.WithContext(ctx).Create(record)

    // 2. Upsert 每日统计 (PostgreSQL ON CONFLICT)
    today := time.Now().Truncate(24 * time.Hour)
    modeColumn := modeToColumn(record.Mode)
    db.WithContext(ctx).Exec(`
        INSERT INTO daily_stats (user_id, date, `+modeColumn+`, total_count, correct_count)
        VALUES (?, ?, 1, 1, ?)
        ON CONFLICT (user_id, date) DO UPDATE SET
            `+modeColumn+` = daily_stats.`+modeColumn+` + 1,
            total_count = daily_stats.total_count + 1,
            correct_count = daily_stats.correct_count + ?,
            updated_at = NOW()
    `, record.UserID, today, correctInc, correctInc)
}

// 在各finish逻辑末尾调用(异步):
go RecordStudyAction(l.ctx, l.svcCtx.Model.DB, &bean.StudyRecord{
    UserID: uid, WordID: wid, Mode: ModeReview, Result: 1,
})
```

#### 连续打卡计算

```go
func GetContinuousDays(ctx context.Context, db *gorm.DB, userID uint) int {
    var dates []time.Time
    db.Model(&bean.DailyStats{}).
        Where("user_id = ? AND total_count > 0", userID).
        Order("date DESC").Limit(365).
        Pluck("date", &dates)

    continuous := 0
    expected := time.Now().Truncate(24 * time.Hour)
    for _, d := range dates {
        dTrunc := d.Truncate(24 * time.Hour)
        if dTrunc.Equal(expected) {
            continuous++
            expected = expected.AddDate(0, 0, -1)
        } else if continuous == 0 && dTrunc.Equal(expected.AddDate(0, 0, -1)) {
            expected = expected.AddDate(0, 0, -1)
            continuous++
            expected = expected.AddDate(0, 0, -1)
        } else {
            break
        }
    }
    return continuous
}
```

#### API设计

```
GET /api/v1/stats/today     → 今日各模式计数+正确率+连续天数
GET /api/v1/stats/calendar?year=2026&month=4  → 每日打卡标记
GET /api/v1/stats/trend?days=30  → 近N天每日学习量
GET /api/v1/stats/overall   → 累计总览(总天数/总词数/总复习/最长连续)
```

#### 性能考虑

- `study_record` 异步goroutine写入，不阻塞主请求
- `daily_stats` UPSERT单条SQL完成增量更新
- 日历查询用 `(user_id, date)` 唯一索引，一个月最多31条
- 连续打卡 `LIMIT 365` + break，效率高
- 历史数据增长到百万级考虑按月分区

#### 影响文件

| 文件 | 变更 |
|------|------|
| `internal/model/bean/study_record.go` | 新增 |
| `internal/model/bean/daily_stats.go` | 新增 |
| `internal/model/bean/schemas.go` | 追加新表AutoMigrate |
| `api/stats.api` | 新增 |
| `api/englishstudy.api` | 追加 import |
| `internal/logic/practise/record.go` | 新增(埋点函数) |
| `internal/logic/stats/*.go` | 新增(统计查询) |
| 所有 finish/submit logic | 末尾追加 RecordStudyAction 调用 |

---

## 全局技术决策汇总

### 设计原则

| 决策 | 共识 |
|------|------|
| **新依赖** | 不引入任何新依赖，图表/日历/进度环全部纯CSS/SVG手写 |
| **组件复用** | 所有练习模式复用 PracticeHeader + LoaderOverlay + CompletionOverlay + usePracticeCards |
| **状态管理** | 保持现有模式(组件级ref/reactive + localStorage)，不引入Pinia |
| **API风格** | 遵循 go-zero goctl 代码生成模式，.api定义 → make api |
| **数据库** | GORM AutoMigrate自动加列，手动执行数据迁移SQL |
| **后端风格** | 新logic手写，handler/types由goctl生成，ServiceContext不变 |

### 模式配色体系

| 模式 | 主色 | 浅色背景 |
|------|------|---------|
| 学习 | `#1989fa` | `#e3f2fd` |
| 复习 | `#ff9800` | `#fff3e0` |
| 强化 | `#ff6b35` | `#fbe9e7` |
| 抽查 | `#7c4dff` | `#ede7f6` |
| 听写(新) | `#00bcd4` | `#e0f7fa` |
| 中译英(新) | `#ab47bc` | `#f3e5f5` |

### 视觉规范

| 设计要素 | 规范 |
|---------|------|
| 页面背景 | `#f8f9fa` |
| 一级页面头部 | `linear-gradient(135deg, ...)` 对应模式色渐变 |
| 二级工具页头部 | 白色 + 底边线 |
| 卡片圆角 | 16px(大)/12px(小)/8px(内嵌) |
| 卡片阴影 | `0 8px 24px rgba(0,0,0,0.08)` |
| 操作按钮高度 | 48px, 圆角 14px |
| 成功色 | `#07c160` |
| 危险色 | `#ee0a24` |
| 正文字号 | 14px |
| 辅助文字 | 12-13px, color `#969799` |

### 新增后端API汇总 (12个)

| # | 方法 | 路径 | 功能 |
|---|------|------|------|
| 1 | GET | `/v1/dashboard` | 今日看板 |
| 2 | GET | `/v1/dictionary/search/word` | 搜索词库 |
| 3 | POST | `/v1/dictionary/search/add` | 快速添加 |
| 4 | GET | `/v1/practise/listen-spell/list` | 听写列表 |
| 5 | POST | `/v1/practise/listen-spell/submit` | 听写提交 |
| 6 | GET | `/v1/practise/cn-to-en/list` | 中译英列表 |
| 7 | POST | `/v1/practise/cn-to-en/submit` | 中译英提交 |
| 8 | GET | `/v1/practise/review/due-count` | 待复习数 |
| 9 | GET | `/v1/stats/today` | 今日统计 |
| 10 | GET | `/v1/stats/calendar` | 打卡日历 |
| 11 | GET | `/v1/stats/trend` | 学习趋势 |
| 12 | GET | `/v1/stats/overall` | 累计总览 |

### 新增前端文件汇总

```
src/
  api/
    dashboard.js              (功能1)
    stardict.js               (功能2)
    stats.js                  (功能6)
  views/
    Dashboard.vue             (功能1)
    SearchWord.vue            (功能2)
    PracticeListening.vue     (功能4)
    PracticeTranslate.vue     (功能5)
    Statistics.vue            (功能6)
  components/
    dashboard/
      TodayTaskCard.vue       (功能1)
      StudyProgressRing.vue   (功能1)
      ModeEntryGrid.vue       (功能1)
    search/
      SearchResultCard.vue    (功能2)
      StardictWordDetail.vue  (功能2)
    practice/
      ReviewScheduleInfo.vue  (功能3)
      SpellingInput.vue       (功能4)
      SpellingResult.vue      (功能4)
    stats/
      WeeklyChart.vue         (功能6)
      CheckinCalendar.vue     (功能6)
      StatsSummaryCard.vue    (功能6)
  composables/
    useSpellingCheck.js       (功能4)
  utils/
    date.js                   (功能3)
```

### 新增后端文件汇总

```
api/
  dashboard.api                   (功能1)
  stats.api                       (功能6)
internal/
  model/bean/
    study_record.go               (功能6)
    daily_stats.go                (功能6)
  model/
    migration.go                  (功能3-SRS迁移)
  logic/
    dashboard/
      getdashboardlogic.go        (功能1)
    dictionary/
      searchwordlogic.go          (功能2)
      quickaddwordlogic.go        (功能2)
    practise/
      srs.go                      (功能3-SM2算法)
      srs_test.go                 (功能3-单测)
      spelling.go                 (功能4+5-通用拼写校验)
      getlistenspelllistlogic.go  (功能4)
      submitlistenspelllogic.go   (功能4)
      getcntoelist*.go            (功能5)
      submitcntoenlogic.go        (功能5)
      record.go                   (功能6-埋点)
    stats/
      gettodaystatslogic.go       (功能6)
      getcalendarstatslogic.go    (功能6)
      gettrendstatslogic.go       (功能6)
      getoverallstatslogic.go     (功能6)
```

### 修改的现有文件汇总

| 文件 | 涉及功能 |
|------|---------|
| `api/englishstudy.api` | 1,6 (import新api文件) |
| `api/dictionary.api` | 2 (追加search路由) |
| `api/practise.api` | 3,4,5 (追加路由+Quality字段) |
| `internal/model/bean/word_status.go` | 3 (添加SRS字段) |
| `internal/model/bean/schemas.go` | 6 (追加新表) |
| `internal/logic/practise/finishstudylogic.go` | 3,6 (SRS初始化+埋点) |
| `internal/logic/practise/finishreviewlogic.go` | 3,6 (SRS计算+埋点) |
| `internal/logic/practise/finishstrengthlogic.go` | 3,6 (SRS参数+埋点) |
| `internal/logic/practise/finishspotlogic.go` | 3,6 (SRS重置+埋点) |
| `internal/logic/practise/getreviewwordcardlistlogic.go` | 3 (查询条件改next_review_at) |
| `internal/logic/practise/getstrengthwordcardlistlogic.go` | 3 (同上) |
| `src/router/index.js` | 1,2,4,5,6 |
| `src/components/TabBar.vue` | 1 |
| `src/views/Practice.vue` | 1,4,5 |
| `src/views/PracticeReview.vue` | 3 |
| `src/views/Dictionary.vue` | 2 |
| `src/views/Profile.vue` | 6 |
| `src/api/practise.js` | 4,5 |
