# English Study (后端)

基于 go-zero 的英语学习平台后端 API。支持词典管理、多种练习模式、AI 内容生成（例句/图片/发音/翻译）、跨用户分享、SRS 自适应间隔重复。

> 前端项目：[english-study-ui](https://github.com/NameOfGoy/english-study-ui)（Vue3 + Vant H5）

## 功能特性

- **词典管理**
  - 单词/短语/词性的 CRUD，Excel 批量导入导出
  - 中文反查 stardict 词库（约 40 万条）+ 批量异步入库
  - stardict 缺失词条时 AI 兜底生成（音标 / 释义 / 词形变化）
  - 标签系统：默认系统标签 + 用户自定义标签
- **多种练习模式**
  - **学习**：新词初学
  - **复习**：SRS 自适应间隔（SM-2 算法），每词每日仅可复习一次，连续通过 5 轮后掌握
  - **加强**：复习不通过的词强化训练，连续 3 轮通过后回到复习
  - **抽查**：随机抽 50 个，按完成时间比例（3 月+ / 1 月+ / 2 周内 = 30:15:5）
- **AI 内容生成**
  - 词例句：智谱 BigModel（GLM-4-Flash）
  - 词图片：智谱 CogView
  - 发音音频：阿里云 NLS TTS
  - 翻译：智谱 LLM
- **用户系统**
  - 账号密码 / 微信 OAuth 登录
  - JWT 鉴权
  - 用户独立词库（每用户独立分表 `word_user_{uid}`、`word_pos_user_{uid}`、`word_phrase_user_{uid}`）
- **跨用户分享**
  - HMAC-SHA256 签名的无状态 token（5 分钟 TTL）
  - 支持按词类型/标签筛选，接收方可预览后选择性导入
- **首页仪表盘**
  - 一接口聚合学习/复习/加强/抽查 + 今日完成数 + 总体进度

## 技术栈

| 类别 | 选型 |
|---|---|
| 框架 | [go-zero](https://github.com/zeromicro/go-zero) v1.8 |
| 语言 | Go 1.23 |
| 数据库 | PostgreSQL + [GORM](https://gorm.io/) v2 + [GORM Gen](https://gorm.io/gen) |
| 对象存储 | MinIO |
| LLM / 图片生成 | 智谱 BigModel (ChatGLM / CogView) |
| 语音合成 | 阿里云 NLS TTS |
| 鉴权 | JWT + 微信 OAuth |
| 配置 | Viper（支持热加载） |
| 代码生成 | goctl |

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
  ├── oss/minio/             # MinIO 封装
  ├── wx/                    # 微信登录
  └── thirdpart/             # 第三方 SDK 封装
docs/
  ├── architecture.md        # 详细架构文档
  └── migrations/            # 数据库迁移 SQL
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
psql "<DSN>" -f docs/migrations/001_srs_fields.sql
psql "<DSN>" -f docs/migrations/002_backfill_next_review_at.sql
# ... 按编号顺序
```

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
| `practise` | `/practise` | 学习 / 复习 / 加强 / 抽查 四种模式 |
| `tag` | `/tag` | 标签 CRUD |
| `dashboard` | `/dashboard` | 首页统计聚合 |
| `share` | `/share` | 跨用户分享生成 / 预览 / 导入 |
| `file-service` | `/file` | 上传 / 下载 / 图片搜索 |

完整 Swagger 见 `api/swagger/englishstudy.yaml`。

## 部署

推荐 Docker + docker-compose。镜像构建：

```bash
make build-backend           # 构建 v0.0.42
docker tag english-study:v0.0.42 <your-registry>/english-study:v0.0.42
docker push <your-registry>/english-study:v0.0.42
```

容器中将配置文件挂到 `/app/etc/englishstudy-api.yaml`，对外暴露 `8888`。

## 相关仓库

- [english-study-ui](https://github.com/NameOfGoy/english-study-ui) — Vue3 + Vant H5 前端

## License

MIT
