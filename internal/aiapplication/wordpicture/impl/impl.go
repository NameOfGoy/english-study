package impl

import (
	"bytes"
	"context"
	"english-study/internal/AI/view"
	"english-study/internal/aiapplication/wordpicture"
	"english-study/internal/config"
	"english-study/internal/utils"
	"text/template"
)

// Generator 是Picture接口的实现
type Generator struct {
	view view.View
	vc   *config.ViperConfig
}

// NewGenerator 创建一个新的Generator
func NewGenerator(v view.View, vc *config.ViperConfig) *Generator {
	return &Generator{
		view: v,
		vc:   vc,
	}
}

// Generate 实现Picture接口的Generate方法
func (g *Generator) Generate(ctx context.Context, word string, opts ...wordpicture.OptionFunc) ([]byte, error) {
	// 处理选项，设置默认值
	opt := utils.GetOptionFromOptions[wordpicture.Option, wordpicture.OptionFunc](opts)
	if opt.PromptTemplate.Template == "" {
		opt.PromptTemplate.Template = wordpicture.DefaultPromptTemplate.Template
	}
	// 准备模板数据
	templateData := wordpicture.TemplateData{
		Word:        word,
		Pos:         opt.Pos,
		Translation: opt.Translation,
	}

	// 解析模板
	tmpl, err := template.New("prompt").Parse(opt.PromptTemplate.Template)
	if err != nil {
		return nil, err
	}

	// 生成提示
	var promptBuf bytes.Buffer
	err = tmpl.Execute(&promptBuf, templateData)
	if err != nil {
		return nil, err
	}

	// 调用AI.View接口生成图片
	vopts := []view.Options{
		func() view.Options {
			if opt.Width != 0 && opt.Height != 0 {
				return view.WithCustomSize(opt.Width, opt.Height)
			}
			return view.WithSize(view.ImageSizeLarge)
		}(),
		view.WithResponseFormat(view.ResponseFormatBase64),
	}
	imageData, err := g.view.Generate(ctx, promptBuf.String(), vopts...)
	if err != nil {
		return nil, err
	}

	return imageData, nil
}
