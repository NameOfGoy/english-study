package svc

import (
	"english-study/internal/aiapplication/articlegen"
	"english-study/internal/aiapplication/wordexample"
	"english-study/internal/aiapplication/wordpicture"
	"english-study/internal/ccauth"
	"english-study/internal/config"
	"english-study/internal/dictionary"
	"english-study/internal/model"
	"english-study/internal/oss"
	"english-study/internal/oss/minio"
	"english-study/internal/wx"

	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config      config.Config
	ViperConfig *config.ViperConfig
	Model       *model.Model
	Oss         oss.OSS
	Wx          *wx.Client
	Dictionary  dictionary.Dictionary
	WordPic     wordpicture.Picture     // 单词图片生成器
	WordExam    wordexample.WordExample // 单词例句生成器
	ArticleGen  articlegen.ArticleGenerator // 短篇文章生成器
	// AI 桥二次门禁: IP 错密钥锁 + refresh token rotation 凭据 store
	// 单实例进程内存; 多实例部署要换 Redis (当前单租户够用)
	CCIPLimiter    *ccauth.IPRateLimiter
	CCRefreshStore *ccauth.RefreshStore
}

type Middleware struct {
	RequestOverwriteMiddleware rest.Middleware
}

func NewServiceContext(c config.Config, vconfig *config.ViperConfig, model *model.Model, minio *minio.Minio, wx *wx.Client, dictionary dictionary.Dictionary, wordPic wordpicture.Picture, wordExam wordexample.WordExample, articleGen articlegen.ArticleGenerator) *ServiceContext {
	return &ServiceContext{
		Config:         c,
		ViperConfig:    vconfig,
		Model:          model,
		Oss:            minio,
		Wx:             wx,
		Dictionary:     dictionary,
		WordPic:        wordPic,
		WordExam:       wordExam,
		ArticleGen:     articleGen,
		CCIPLimiter:    ccauth.NewIPRateLimiter(5, 5*60, 5*60), // 5 次/5min 窗口, 锁 5min
		CCRefreshStore: ccauth.NewRefreshStore(),
	}
}
