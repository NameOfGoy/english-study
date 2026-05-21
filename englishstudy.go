package main

import (
	"english-study/internal/AI/llm/bigmodel"
	"english-study/internal/AI/tts/alicloud"
	viewbigmodel "english-study/internal/AI/view/bigmodel"
	englishwordexampleimpl "english-study/internal/aiapplication/wordexample/impl"
	wordpictureimpl "english-study/internal/aiapplication/wordpicture/impl"
	wordpronounceimpl "english-study/internal/aiapplication/wordpronounce/impl"
	wordtranslationimpl "english-study/internal/aiapplication/wordtranslation/impl"
	dictionaryimpl "english-study/internal/dictionary/impl"
	"english-study/internal/model"
	"english-study/internal/oss/minio"
	"english-study/internal/thirdpart/alibabacloud"
	"english-study/internal/wx"
	"flag"
	"fmt"
	"log"

	"english-study/internal/config"
	"english-study/internal/handler"
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

	dict := dictionaryimpl.NewDictionaryImpl(mi, m, exampleGenerator, wordPicture, wordPronounce, wordTranslation)
	// 进程退出前等待后台例句生成等异步任务完成, 避免半写状态
	defer dict.WaitBackground()

	ctx := svc.NewServiceContext(c, vc, m, mi, wxc, dict, wordPicture, exampleGenerator)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
