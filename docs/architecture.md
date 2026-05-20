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
- 数据库: PostgreSQL + GORM
- 对象存储: MinIO
- AI: 智谱BigModel(LLM+图片生成) + 阿里云NLS(TTS)
- 认证: JWT + 微信OAuth

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
├── oss/                     # MinIO封装
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
| POST | /operation/import | 批量导入(异步) |
| POST | /operation/export | 导出 |

#### 练习 `/api/v1/practise`
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /study/list | 获取学习单词卡 |
| POST | /study/finish | 完成学习 |
| GET | /review/list | 获取复习单词卡(仅StudyTime<today) |
| POST | /review/finish | 复习完成/失败 |
| GET | /strength/list | 获取加强记忆单词卡 |
| POST | /strength/finish | 加强完成/失败 |
| GET | /spot/list | 获取抽查单词卡 |
| POST | /spot/finish | 抽查通过/失败 |

### 3.4 数据模型

**全局表:**
- `users` - 用户(ID, Account, WxOpenID, Username, Password, Phone, Email, Avatar)
- `word` - 词典单词(ID, Word, 美音音标, 美音音频, 英音音标, 英音音频)
- `word_pos` - 词性(ID, WordID, Word, Pos, Translation, Example[JSON], Picture, Exchange[JSON])
- `word_phrase` - 短语(ID, Phrase, Translation, Pronunciation, Example[JSON], Picture)
- `stardict` - 标准词典(外部数据源)
- `tags` - 标签(ID, Tag, Style, UserID)

**用户表(动态创建 `{table}_user_{userId}`):**
- `word_user_{uid}` - 用户单词副本
- `word_pos_user_{uid}` - 用户词性副本
- `word_phrase_user_{uid}` - 用户短语副本

**关联表:**
- `word_status` - 学习状态(WordID, WordType, Status, Times, Weight, StudyTime, UserID)
- `word_tag` - 单词标签关联(WordID, WordType, TagID, UserID)

**WordType:** 1=单词, 2=短语
**Pos枚举:** 1-n, 2-vt, 3-vi, 4-adv, 5-adj, 6-prep, 7-conj, 8-interj, 9-part, 10-pron, 11-num, 12-art, 13-aux

### 3.5 AI服务架构

```
aiapplication/
├── wordexample/     # 例句生成 → BigModel LLM (JSON输出)
├── wordpicture/     # 图片生成 → BigModel Image API → MinIO
├── wordpronounce/   # 发音生成 → 阿里云NLS TTS → MinIO
└── wordtranslation/ # 翻译 → BigModel LLM
```

### 3.6 关键实现

- **单词导入:** 异步goroutine处理，逐行读文件，自动查StarDict、生成语音/例句/图片
- **并发控制:** Dictionary用RWMutex+map防止重复并发添加同一单词
- **密码:** PBKDF2 (2048迭代, SHA256, 32字节key)
- **MinIO路径:** `picture/user_word_{uid}/{word}/{pos}_{ts}.png`, `pronounce/{word}/{accent}.mp3`

## 4. 前端架构

### 4.1 技术栈
Vue 3.3.4 + Vite + Vue Router + Axios + Vant 4.6.2 + SCSS

### 4.2 目录结构
```
English-Study-UI/src/
├── api/          # user.js, dictionary.js, practise.js, tag.js, file.js
├── components/   # TabBar.vue, SearchModal.vue
├── router/       # 路由配置
├── utils/        # auth.js(localStorage管理), request.js(Axios封装)
├── views/        # Login, Register, Profile, Practice(hub+4模式), Dictionary
├── styles/       # SCSS变量+全局样式
└── main.js
```

### 4.3 路由
- `/` → `/practice` (默认首页)
- `/login`, `/register` (公开)
- `/practice` (练习入口, 4个模式卡片)
- `/practice/study|review|strength|spot` (4种练习页面)
- `/dictionary` (词典, 单词/短语tab)
- `/profile` (个人中心)

### 4.4 页面交互

**PracticeStudy(学习):** 图片轮播 + 单词 + 英美发音 + 翻译 + 例句轮播 → "完成学习"
**PracticeReview(复习):** 3阶段显示(图片→提示→完整) → "已掌握"/"未掌握"
**PracticeStrength(加强):** 显示中文 → "记得"/"不记得" → 进度条
**PracticeSpot(抽查):** 类似复习流程
**Dictionary:** 3种视图(字母/状态/标签) + 搜索 + 词/短语切换

### 4.5 部署
Docker + Nginx, dev代理 `/api` → `localhost:8888`

## 5. 前后端联调

- 前端 baseURL: `/api/v1`
- 认证: JWT Bearer token
- 401 → 自动跳登录页
- 资源URL: 开发环境用远程MinIO `http://193.112.111.2:39000/`
