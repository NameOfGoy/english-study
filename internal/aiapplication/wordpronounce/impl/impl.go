package impl

import (
	"bytes"
	"context"
	"english-study/internal/AI/llm"
	"english-study/internal/aiapplication/wordpronounce"
	"english-study/internal/config"
	"english-study/internal/utils"
	"text/template"

	"english-study/internal/AI/tts"
)

// Generator 发音生成器
type Generator struct {
	tts tts.TTS
	llm llm.LLM
	vc  *config.ViperConfig
}

// NewGenerator 创建新的发音生成器
func NewGenerator(tts tts.TTS, llm llm.LLM, vc *config.ViperConfig) *Generator {
	return &Generator{
		tts: tts,
		llm: llm,
		vc:  vc,
	}
}

// Generate 生成语音
func (g *Generator) GeneratePronounce(ctx context.Context, word string, opts ...wordpronounce.OptionFunc) ([]byte, error) {
	// 使用 utils.GetOptionFromOptions 处理选项，设置默认值
	opt := utils.GetOptionFromOptions(opts)
	// 设置选项
	options := []tts.Option{
		tts.WithAccent(opt.Accent),
		tts.WithSpeechRate(-50),
		tts.WithFormat("mp3"),
	}
	// 调用 TTS 接口生成语音
	return g.tts.TextToSpeech(ctx, word, options...)
}

func (g *Generator) GeneratePronouncePhonetic(ctx context.Context, word string, opts ...wordpronounce.OptionFunc) (string, error) {
	// 处理选项，设置默认值
	opt := utils.GetOptionFromOptions(opts)

	// 从配置文件获取模板
	promptTemplateStr := g.vc.GetWordPhoneticPromptTemplate()

	// 准备模板数据
	templateData := wordpronounce.TemplateData{
		Word:   word,
		Accent: opt.Accent,
	}

	// 解析模板
	tmpl, err := template.New("phonetic").Parse(promptTemplateStr)
	if err != nil {
		return "", err
	}

	// 生成提示
	var promptBuf bytes.Buffer
	err = tmpl.Execute(&promptBuf, templateData)
	if err != nil {
		return "", err
	}

	// 调用LLM生成音标
	response, err := g.llm.Chat(ctx, promptBuf.String())
	if err != nil {
		return "", err
	}

	return response, nil
}
