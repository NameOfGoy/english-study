package impl

import (
	"bytes"
	"context"
	"encoding/json"
	"english-study/internal/AI/llm"
	"english-study/internal/aiapplication/wordtranslation"
	"english-study/internal/utils"
	"fmt"
	"strings"
	"text/template"
)

// Generator 是 WordTranslation 接口的实现
type Generator struct {
	llm            llm.LLM
	promptTemplate wordtranslation.PromptTemplate
}

// NewGenerator 创建一个新的 Generator
func NewGenerator(l llm.LLM) *Generator {
	return &Generator{
		llm:            l,
		promptTemplate: wordtranslation.DefaultPromptTemplate,
	}
}

// WithPromptTemplate 设置自定义提示模板
func (g *Generator) WithPromptTemplate(t wordtranslation.PromptTemplate) *Generator {
	g.promptTemplate = t
	return g
}

// Generate 根据词语生成中文翻译
func (g *Generator) Generate(ctx context.Context, word string, opts ...wordtranslation.OptionFunc) (translation string, err error) {
	// 处理选项，设置默认值
	opt := utils.GetOptionFromOptions[wordtranslation.Option, wordtranslation.OptionFunc](opts, func() wordtranslation.Option {
		return wordtranslation.Option{}
	})

	// 准备模板数据
	tmplData := wordtranslation.TemplateData{
		Word: word,
		Pos:  opt.Pos,
	}

	// 生成提示词
	tmpl, err := template.New("prompt").Parse(g.promptTemplate.Template)
	if err != nil {
		return "", err
	}

	var promptBuf bytes.Buffer
	if err := tmpl.Execute(&promptBuf, tmplData); err != nil {
		return "", err
	}

	// 调用 LLM 生成翻译
	answer, err := g.llm.Chat(ctx, promptBuf.String())
	if err != nil {
		return "", err
	}

	// 去除可能的markdown代码块修饰并裁剪空白，直接返回纯文本
	plain := strings.TrimSpace(utils.TrimMarkDownJsonTag(answer))

	if strings.Contains(plain, "错误：") {
		return "", fmt.Errorf(plain)
	}
	return plain, nil
}

// wordInfoPrompt 让 LLM 输出严格 JSON 的 prompt
const wordInfoPrompt = `你是一个英语词典助手。请为英语单词 "%s" 生成完整的词典条目信息。

要求严格按以下 JSON 格式输出，不要包含任何 markdown 标记或额外说明：

{
  "valid": true,
  "phonetic": "/...../",
  "translation": "n. 中文释义\nv. 中文释义",
  "definition": "English definition here",
  "exchange": "p:past,d:past_part,i:present_part,3:third_singular,s:plural"
}

字段说明：
- valid: 如果不是合法英语单词、拼写错误、或非英语词汇，返回 {"valid": false, "reason": "原因"}
- phonetic: 国际音标，必须包含 / / 包裹符
- translation: 多行格式，每行 "词性. 中文释义"。词性缩写从这些里选: n. vt. vi. v. adj. adv. prep. conj. interj. pron. num. art.
- definition: 英文释义，简洁 1-2 句
- exchange: 词形变化，逗号分隔。键: p(过去式) d(过去分词) i(现在分词) 3(三单) r(比较级) t(最高级) s(复数)。该单词没有的变化形式可省略键。如果完全没有变化（如形容词无比较级），返回空字符串 ""

只返回 JSON，不要 ` + "```json" + ` 包裹，不要解释。`

// GenerateWordInfo 让 LLM 充当 stardict 兜底
func (g *Generator) GenerateWordInfo(ctx context.Context, word string) (*wordtranslation.WordInfo, error) {
	prompt := fmt.Sprintf(wordInfoPrompt, word)
	answer, err := g.llm.Chat(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM 调用失败: %w", err)
	}

	// 清理可能的 markdown 包裹
	cleaned := strings.TrimSpace(utils.TrimMarkDownJsonTag(answer))

	var info wordtranslation.WordInfo
	if err := json.Unmarshal([]byte(cleaned), &info); err != nil {
		// 尝试找到 JSON 部分（防 LLM 多输出）
		start := strings.Index(cleaned, "{")
		end := strings.LastIndex(cleaned, "}")
		if start < 0 || end <= start {
			return nil, fmt.Errorf("AI 返回不是有效 JSON: %s", truncateForLog(cleaned))
		}
		if err := json.Unmarshal([]byte(cleaned[start:end+1]), &info); err != nil {
			return nil, fmt.Errorf("AI 返回 JSON 解析失败: %w (原文: %s)", err, truncateForLog(cleaned))
		}
	}

	if !info.Valid {
		if info.Reason == "" {
			info.Reason = "未识别为合法英语单词"
		}
		return &info, nil
	}

	// 基本校验：合法时必须有翻译和音标
	if strings.TrimSpace(info.Translation) == "" {
		return nil, fmt.Errorf("AI 返回的翻译为空")
	}
	if strings.TrimSpace(info.Phonetic) == "" {
		// 音标缺失不是致命错误，留空让下游兜底
		info.Phonetic = ""
	}

	return &info, nil
}

func truncateForLog(s string) string {
	const maxLen = 200
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
