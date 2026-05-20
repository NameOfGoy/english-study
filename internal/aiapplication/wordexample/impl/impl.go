package impl

import (
	"bytes"
	"context"
	"encoding/json"
	"english-study/internal/AI/llm"
	"english-study/internal/aiapplication/wordexample"
	"english-study/internal/types"
	"english-study/internal/utils"
	"text/template"
)

type Generator struct {
	llm            llm.LLM
	promptTemplate wordexample.PromptTemplate
}

func NewGenerator(l llm.LLM) *Generator {
	return &Generator{
		llm:            l,
		promptTemplate: wordexample.DefaultPromptTemplate,
	}
}

// WithPromptTemplate 设置自定义提示模板
func (g *Generator) WithPromptTemplate(template wordexample.PromptTemplate) *Generator {
	g.promptTemplate = template
	return g
}

func (g *Generator) Generate(ctx context.Context, word string, opts ...wordexample.OptionFunc) (exams []*types.Example, err error) {
	opt := utils.GetOptionFromOptions[wordexample.Option, wordexample.OptionFunc](opts, func() wordexample.Option {
		return wordexample.Option{
			Count: 1,
		}
	})

	// 准备模板数据
	tmplData := wordexample.TemplateData{
		Word:        word,
		Pos:         opt.Pos,
		Count:       opt.Count,
		Translation: opt.Translation,
		JSONFormat:  g.promptTemplate.JSONFormat,
	}

	// 使用模板生成提示
	tmpl, err := template.New("prompt").Parse(g.promptTemplate.Template)
	if err != nil {
		return nil, err
	}

	var promptBuf bytes.Buffer
	if err := tmpl.Execute(&promptBuf, tmplData); err != nil {
		return nil, err
	}

	// 调用LLM生成例句
	answer, err := g.llm.Chat(ctx, promptBuf.String())
	if err != nil {
		return nil, err
	}

	// 提取ai回复的json字符串
	jsonStr := utils.TrimMarkDownJsonTag(answer)

	// 解析JSON响应
	var es []*types.Example
	err = json.Unmarshal([]byte(jsonStr), &es)
	if err != nil {
		return nil, err
	}

	return es, nil
}
