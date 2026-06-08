package config

import "github.com/zeromicro/go-zero/rest"

// ArticleGenConf 短篇文章 AI 生成配置. 具名类型以便跨包传给 articlegen.NewGenerator.
// 字段均 optional, yaml 缺省时为零值, 生成器代码层兜底默认.
type ArticleGenConf struct {
	Model      string `json:",optional"` // 生成所用模型, 空则用 llm 默认免费模型
	RetryCount int    `json:",optional"` // 校验失败纠错重问次数, <=0 时 main 兜底为 1
}

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	Postgresql struct {
		DSN string
	}
	Minio struct {
		Endpoint  string
		AccessKey string
		SecretKey string
		UseSSL    bool
		Domain    string
		Bucket    string
	}
	WxApp struct {
		AppID     string
		AppSecret string
	}
	AliCloud struct {
		AccessKeyId     string
		AccessKeySecret string
		Region          string
		AppKey          string
	}
	BigModel struct {
		APIKey  string
		Article ArticleGenConf `json:",optional"` // 文章生成配置, 缺省即全零值(走代码默认)
	}
	// CC AI 桥 (前端→转发器→本地 Claude Code) 配置.
	//
	// 鉴权策略 (二次门禁):
	//   1) 第一道: english-study 主 JWT (admin role 才有 /admin/chat 入口)
	//   2) 第二道: 用户进 /admin/chat 时再输 AccessKey, 后端比对 + 签发 access + refresh token
	//
	// AccessKey 跟主登录密码物理隔离, 防止"admin 登录态被拿到就直接进 AI 桥"; 也对个人 admin
	// 自己: token 短 TTL + refresh rotation, 任一被盗损失面收敛.
	CC struct {
		RelayWSURL       string // wss://<domain>/forwarder/ws, 给前端用
		AccessKey        string // 进 /admin/chat 第二道关密钥, 跟主登录密码不一样
		AccessTokenTTL   int64  // access token 有效期 (秒), 默认 900 (15min); 转发器 ws 接受
		RefreshTokenTTL  int64  // refresh token 有效期 (秒), 默认 604800 (7d); 仅 /relay-refresh 接受
	}
}
