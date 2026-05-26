# English Study (后端)

基于 go-zero 的英语学习平台后端 API。支持词典管理、多种练习模式、AI 内容生成（例句/图片/发音/翻译）、跨用户分享、SRS 自适应间隔重复。

> 前端项目：[english-study-ui](https://github.com/NameOfGoy/english-study-ui)（Vue3 + Vant H5）

## 功能特性

- **词典管理**
  - 单词/短语/词性的 CRUD（含删除级联清理学习状态与标签关联），文本批量导入
  - 中文反查 stardict 词库（约 40 万条）+ 批量异步入库
  - stardict 缺失词条时 AI 兜底生成（音标 / 释义 / 词形变化）
  - **导入文件支持标签语法**：`[tag]` 段标签 / `[---]` 终止符 / 行末内联 `apple [t1] [t2]` 多标签，详见 [`docs/import-format.txt`](docs/import-format.txt)
- **标签系统（分级）**
  - **系统标签**（`tag.user_id=0`）：所有用户可用，**仅超管可增删改**
  - **用户标签**：按用户隔离，本人增删改
  - 同名优先复用：导入解析时系统标签 > 用户已有 > 自动新建用户级
  - 删除采用进程内事件总线 [resource-eventbus](https://github.com/NameOfGoy/resource-eventbus) 异步级联清 `word_tags`，启动孤儿 reaper 兜底
- **多种练习模式 + 全局标签筛选**
  - **学习**：新词初学
  - **复习**：SRS 自适应间隔（SM-2 算法），每词每日仅可复习一次，连续通过 5 轮后掌握
  - **加强**：复习不通过的词强化训练，连续 3 轮通过后回到复习
  - **抽查**：随机抽 50 个，按完成时间比例（3 月+ / 1 月+ / 2 周内 = 30:15:5）
  - **TagFilterBar**：练习页顶部多选标签，跨 4 模式生效，localStorage 持久化
- **AI 内容生成**
  - 词例句：智谱 BigModel（GLM-4-Flash）
  - 词图片：智谱 CogView
  - 发音音频：阿里云 NLS TTS
  - 翻译：智谱 LLM
- **用户系统 / 权限**
  - 账号密码 / 微信 OAuth 登录
  - 密码：bcrypt(cost=12)，兼容历史 PBKDF2 + 明文，登录成功自动升级
  - JWT (HMAC-SHA256) 携带 `user_id/username/role` 三个 claim；服务端鉴权一律读 JWT, 不信任请求体的 role
  - User.Role：0=普通 / 1=超管；超管可管理系统标签
  - 用户独立词库（每用户独立分表 `word_user_{uid}`、`word_pos_user_{uid}`、`word_phrase_user_{uid}`）
- **跨用户分享**
  - HMAC-SHA256 全 32 字节签名 + 64-bit nonce 的无状态 token（5 分钟 TTL）
  - 支持按词类型/标签筛选；导入端三种模式：不带标签 / 仅系统标签 / 全部
- **首页仪表盘**
  - 一接口聚合学习/复习/加强/抽查 + 今日完成数 + 总体进度
- **安全加固**
  - 全局 `SecurityHeaders` 中间件（X-Content-Type-Options / X-Frame-Options / Referrer-Policy）
  - proxyImage：白名单 + Transport.DialContext 在 dial 时验证 IP，闭合 DNS rebinding；io.Copy 限到 20MB
  - OSS 对象 key 全部走 `oss.SafeKeyPart` 防 `../` 越权

## 技术栈

| 类别 | 选型 |
|---|---|
| 框架 | [go-zero](https://github.com/zeromicro/go-zero) v1.8 |
| 语言 | Go 1.23 |
| 数据库 | PostgreSQL + [GORM](https://gorm.io/) v2 + [GORM Gen](https://gorm.io/gen) |
| 对象存储 | MinIO |
| LLM / 图片生成 | 智谱 BigModel (ChatGLM / CogView) |
| 语音合成 | 阿里云 NLS TTS |
| 鉴权 | JWT (含 role claim) + bcrypt + 微信 OAuth |
| 事件总线 | [resource-eventbus](https://github.com/NameOfGoy/resource-eventbus) (in-process typed Pub/Sub) |
| 配置 | Viper（支持热加载） |
| 代码生成 | goctl + GORM Gen |

## 目录结构

```
englishstudy.go              # 入口：初始化 DB/MinIO/微信/AI/路由
api/                         # API 定义（goctl 格式）
  ├── dashboard.api          # 首页仪表盘
  ├── dictionary.api         # 词典 CRUD/导入/搜索
  ├── englishstudy.api       # 入口聚合
  ├── fileservice.api        # 文件上传下载
  ├── practise.api           # 四种练习模式
  ├── share.api              # 跨用户分享
  ├── tag.api                # 标签
  └── user.api               # 用户登录注册
internal/
  ├── handler/               # goctl 生成的 HTTP 处理器
  ├── logic/                 # 业务逻辑（手写）
  │   ├── dashboard/
  │   ├── dictionary/
  │   ├── practise/
  │   │   ├── srs.go         # SM-2 自适应间隔算法
  │   │   └── utils_status_transfer.go  # 单词状态机
  │   ├── share/
  │   ├── tag/
  │   ├── user/
  │   └── file/
  ├── svc/                   # 服务上下文（DI 容器）
  ├── config/                # 配置结构体
  ├── model/
  │   ├── bean/              # GORM 实体
  │   └── dto/               # GORM Gen 生成的查询代码
  ├── dictionary/            # 词典接口 + 实现（含 AI 兜底）
  ├── AI/
  │   ├── llm/bigmodel/      # 智谱 LLM
  │   ├── tts/alicloud/      # 阿里云 TTS
  │   └── view/bigmodel/     # 智谱图片生成
  ├── aiapplication/         # AI 应用封装（例句/图片/发音/翻译）
  ├── eventbus/              # resource-eventbus 集成: payload + subscriber + 启动 reaper
  ├── handler/middleware/    # SecurityHeaders 中间件
  ├── oss/minio/             # MinIO 封装 (含 SafeKeyPart 防路径越权)
  ├── wx/                    # 微信登录
  └── thirdpart/             # 第三方 SDK 封装
docs/
  ├── architecture.md        # 详细架构文档
  ├── import-format.txt      # 导入文件文本格式说明 (面向终端用户)
  ├── tag-overhaul-design.md # 标签系统改造设计文档
  └── migrations/            # 数据库迁移 SQL (001-009)
```

## 快速开始

### 前置要求

- Go 1.23+
- PostgreSQL 13+
- MinIO（或兼容 S3 协议的对象存储）
- 智谱 BigModel API Key（[申请](https://open.bigmodel.cn/)）
- 阿里云 NLS AppKey 与 RAM AccessKey（用于 TTS）
- 微信小程序 AppID / AppSecret（如需微信登录）
- `goctl`（如需重新生成代码）：`go install github.com/zeromicro/go-zero/tools/goctl@latest`

### 配置文件

复制示例并按实际填写：

```bash
mkdir -p etc
cat > etc/englishstudy-api.yaml <<'EOF'
Name: englishstudy-api
Host: 0.0.0.0
Port: 8888
Timeout: 300000

Auth:
  AccessSecret: "<你的-JWT-签名密钥-请用随机长串>"
  AccessExpire: 86400

Postgresql:
  DSN: "postgres://<user>:<password>@<host>:<port>/<dbname>?sslmode=disable"

Minio:
  Endpoint: "<minio-host>:<port>"
  AccessKey: "<your-minio-key>"
  SecretKey: "<your-minio-secret>"
  UseSSL: false
  Domain: "http://<minio-host>:<port>"
  Bucket: "englishstudy"

WxApp:
  AppID: "<wx-app-id>"
  AppSecret: "<wx-app-secret>"

AliCloud:
  AccessKeyId: "<aliyun-ram-access-key-id>"
  AccessKeySecret: "<aliyun-ram-access-key-secret>"
  Region: ""
  AppKey: "<aliyun-nls-app-key>"

BigModel:
  APIKey: "<bigmodel-api-key>"
  BaseURL: "https://open.bigmodel.cn/api/paas/v4"
  Timeout: 30000

# Prompt 模板（按需修改）
WordPicturePromptTemplate: |
  请为英语单词'{{.Word}}'... （详见示例）
EOF
```

> **该文件已在 `.gitignore` 中，不会被提交。**

### 数据库初始化

服务启动时通过 GORM `AutoMigrate` 自动建表。如已有库迁移，按需执行 `docs/migrations/` 下的 SQL 脚本：

```bash
for f in docs/migrations/0*_*.sql; do
  psql "<DSN>" -f "$f"
done
```

迁移说明:
- `001` SRS 字段; `002` 回填 next_review_at
- `003` 默认 8 个系统标签 (`user_id=0`)
- `004` 修 null example; `005` 用户表例句回填
- `006` 热路径索引 (pg_trgm GIN, 复合索引等)
- `007` word_statuses / tags 唯一约束
- `008` users.role 字段, `id=0` 升超管 (按需调整或自行指定其它账号)
- `009` word_tags 复合索引 (给练习页全局标签筛选用)

### 启动

```bash
# 1. 拉依赖
go mod tidy

# 2. 启动服务（默认端口 8888）
go run englishstudy.go -f etc/englishstudy-api.yaml
```

服务启动后会自动在 stardict 库不存在时初始化（首次启动较慢）。

### 常用 Make 命令

```bash
make api         # 修改 .api 后重新生成 handler/types
make fmt         # 格式化 .api 文件
make doc         # 生成 Swagger 到 api/swagger/
make error       # 从 proto 生成错误类型
make build-backend  # 构建后端 Docker 镜像
```

## API 端点概览

所有路由前缀 `/api/v1/`，除登录/注册外都需 JWT。

| 分组 | 路由前缀 | 主要功能 |
|---|---|---|
| `user` | `/user` | 登录、注册（账密 + 微信）、用户信息 |
| `dictionary` | `/dictionary` | 单词/短语 CRUD、导入导出、stardict 搜索 |
| `practise` | `/practise` | 学习 / 复习 / 加强 / 抽查 四种模式（list 接 `tag_ids` 做全局筛选）|
| `tag` | `/tag` | 标签 CRUD（系统标签仅超管可改）|
| `dashboard` | `/dashboard` | 首页统计聚合 |
| `share` | `/share` | 跨用户分享生成 / 预览 / 导入（3 种标签模式）|
| `file-service` | `/file` | 上传 / 下载 / 图片搜索 / 代理图片（带 SSRF 防御）|

完整 Swagger 见 `api/swagger/englishstudy.yaml`。

## 部署

推荐 Docker + docker-compose。镜像构建：

```bash
make build-backend           # 当前版本见 Makefile 的 VERSION
docker tag english-study:<ver> <your-registry>/english-study:<ver>
docker push <your-registry>/english-study:<ver>
```

容器中将配置文件挂到 `/app/etc/englishstudy-api.yaml`，对外暴露 `8888`。

## 相关仓库

- [english-study-ui](https://github.com/NameOfGoy/english-study-ui) — Vue3 + Vant H5 前端
- [resource-eventbus](https://github.com/NameOfGoy/resource-eventbus) — 进程内 typed Pub/Sub + DSL 代码生成 (本项目用作 Tag 删除等领域事件的总线)

## License

MIT
