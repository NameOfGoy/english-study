package svc

import (
	"english-study/internal/aiapplication/wordexample"
	"english-study/internal/aiapplication/wordpicture"
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
}

type Middleware struct {
	RequestOverwriteMiddleware rest.Middleware
}

func NewServiceContext(c config.Config, vconfig *config.ViperConfig, model *model.Model, minio *minio.Minio, wx *wx.Client, dictionary dictionary.Dictionary, wordPic wordpicture.Picture, wordExam wordexample.WordExample) *ServiceContext {
	return &ServiceContext{
		Config:      c,
		ViperConfig: vconfig,
		Model:       model,
		Oss:         minio,
		Wx:          wx,
		Dictionary:  dictionary,
		WordPic:     wordPic,
		WordExam:    wordExam,
	}
}
