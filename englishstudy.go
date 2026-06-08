package main

import (
	"english-study/internal/AI/llm/bigmodel"
	"english-study/internal/AI/tts/alicloud"
	viewbigmodel "english-study/internal/AI/view/bigmodel"
	"english-study/internal/aiapplication/articlegen"
	articlegenimpl "english-study/internal/aiapplication/articlegen/impl"
	englishwordexampleimpl "english-study/internal/aiapplication/wordexample/impl"
	wordpictureimpl "english-study/internal/aiapplication/wordpicture/impl"
	wordpronounceimpl "english-study/internal/aiapplication/wordpronounce/impl"
	wordtranslationimpl "english-study/internal/aiapplication/wordtranslation/impl"
	dictionaryimpl "english-study/internal/dictionary/impl"
	cchandler "english-study/internal/handler/cc"
	dictlogic "english-study/internal/logic/dictionary"
	"english-study/internal/model"
	"english-study/internal/oss/minio"
	"english-study/internal/thirdpart/alibabacloud"
	"english-study/internal/wx"
	"flag"
	"fmt"
	"log"
	"net/http"

	"english-study/internal/config"
	"english-study/internal/handler"
	"english-study/internal/handler/middleware"
	"english-study/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/englishstudy-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 初始化Viper配置管理器 - 使用合并路径方式
	vc, err := config.NewViperConfig(*configFile, "")
	if err != nil {
		log.Fatalf("初始化Viper配置失败: %v", err)
	}
	vc.WatchConfig(nil)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	m, err := model.NewModel(c.Postgresql.DSN)
	if err != nil {
		panic(err)
	}

	// 事件总线启动钩子: 1) 兜底清掉上次进程 PublishAsync 后崩了导致的 word_tags 残留;
	// 2) 挂载 tag.deleted 订阅, 后续 tag 删除时异步级联清 word_tags.
	// 放在 main 而非 model 包是为了避免 model → logic → svc → model 循环依赖.
	if err := dictlogic.ReapOrphanWordTags(m.DB); err != nil {
		panic(err)
	}
	if err := dictlogic.SubscribeTagEvents(m.DB); err != nil {
		panic(err)
	}

	mi, err := minio.NewMinio(c.Minio.Endpoint, c.Minio.AccessKey, c.Minio.SecretKey, c.Minio.UseSSL)
	if err != nil {
		panic(err)
	}

	wxc := wx.NewWxClient(c.WxApp.AppID, c.WxApp.AppSecret)

	// 创建TTS实例（阿里云）
	ttsConfig := &alibabacloud.TTSConfig{
		AccessKeyID:     c.AliCloud.AccessKeyId,
		AccessKeySecret: c.AliCloud.AccessKeySecret,
		Region:          c.AliCloud.Region,
		AppKey:          c.AliCloud.AppKey,
	}
	ttsInstance := alicloud.NewAliCloudTTS(ttsConfig)

	// 创建LLM实例（BigModel）
	llmInstance := bigmodel.NewBigModelLLM(c.BigModel.APIKey)

	// 创建View实例（BigModel）
	viewInstance := viewbigmodel.NewBigModelView(c.BigModel.APIKey)

	// 创建aiapplication层实例
	exampleGenerator := englishwordexampleimpl.NewGenerator(llmInstance)
	wordPronounce := wordpronounceimpl.NewGenerator(ttsInstance, llmInstance, vc)
	wordPicture := wordpictureimpl.NewGenerator(viewInstance, vc)
	wordTranslation := wordtranslationimpl.NewGenerator(llmInstance)

	// 文章生成器: yaml 未配 RetryCount(零值) 时兜底为 1(校验失败纠错重问一次, 总调用<=2)
	articleRetry := c.BigModel.Article.RetryCount
	if articleRetry <= 0 {
		articleRetry = 1
	}
	articleGenerator := articlegenimpl.NewGenerator(llmInstance, articlegen.Config{
		Model:      c.BigModel.Article.Model,
		RetryCount: articleRetry,
	})

	dict := dictionaryimpl.NewDictionaryImpl(mi, m, exampleGenerator, wordPicture, wordPronounce, wordTranslation)
	// 进程退出前等待后台例句生成等异步任务完成, 避免半写状态
	defer dict.WaitBackground()

	ctx := svc.NewServiceContext(c, vc, m, mi, wxc, dict, wordPicture, exampleGenerator, articleGenerator)
	// 全局给所有响应加安全头 (X-Content-Type-Options 等)
	server.Use(middleware.SecurityHeaders)
	handler.RegisterHandlers(server, ctx)

	// AI 桥文件传输的两个路由手动注册 (不走 goctl): 上传是 multipart, 下载是流式 + 自定义 token 鉴权,
	// 都不适合 goctl typed handler; 手写还能避免 make api 重新生成时被覆盖.
	registerCCFileRoutes(server, ctx, c.Auth.AccessSecret)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

// registerCCFileRoutes 注册 AI 桥文件上传/下载.
//   - upload: 走 JWT 鉴权 (handler 内再 RequireAdmin); 上传到私有桶
//   - download: 无 JWT 中间件, 凭 ?t= 签名 token 自鉴权 (CC 本地凭 token 下载, 不需业务 JWT)
func registerCCFileRoutes(server *rest.Server, ctx *svc.ServiceContext, jwtSecret string) {
	server.AddRoutes(
		[]rest.Route{
			{Method: http.MethodPost, Path: "/upload-file", Handler: cchandler.UploadFileHandler(ctx)},
		},
		rest.WithJwt(jwtSecret),
		rest.WithPrefix("/api/v1/cc"),
	)
	server.AddRoutes(
		[]rest.Route{
			{Method: http.MethodGet, Path: "/download", Handler: cchandler.DownloadFileHandler(ctx)},
		},
		rest.WithPrefix("/api/v1/cc"),
	)
}
