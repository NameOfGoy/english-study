package impl

import (
	"bytes"
	"context"
	"encoding/json"
	"english-study/internal/AI/llm"
	"english-study/internal/aiapplication/articlegen"
	"english-study/internal/errors"
	"english-study/internal/types"
	"english-study/internal/utils"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

// Generator 是 ArticleGenerator 接口的实现.
type Generator struct {
	llm            llm.LLM
	promptTemplate articlegen.PromptTemplate
	cfg            articlegen.Config
}

// NewGenerator 创建文章生成器. cfg 来自 config.BigModel.Article(由 main 转换).
func NewGenerator(l llm.LLM, cfg articlegen.Config) *Generator {
	if cfg.RetryCount < 0 {
		cfg.RetryCount = 0
	}
	return &Generator{
		llm:            l,
		promptTemplate: articlegen.DefaultPromptTemplate,
		cfg:            cfg,
	}
}

// WithPromptTemplate 设置自定义提示模板.
func (g *Generator) WithPromptTemplate(t articlegen.PromptTemplate) *Generator {
	g.promptTemplate = t
	return g
}

func (g *Generator) Generate(ctx context.Context, words []articlegen.InputWord, opts ...articlegen.OptionFunc) (*articlegen.Article, error) {
	opt := utils.GetOptionFromOptions[articlegen.Option, articlegen.OptionFunc](opts, func() articlegen.Option {
		return articlegen.Option{Model: g.cfg.Model, RetryCount: g.cfg.RetryCount}
	})
	count := len(words)
	if count == 0 {
		return nil, errors.ErrorRequestParamError("生成文章的词条为空")
	}

	// 句数明显多于词数, 给"过渡句"与"一句多词"留预算, 从结构上打破"一词一句".
	minSent := count + 2
	if minSent < 6 {
		minSent = 6
	}
	maxSent := count + 5
	if maxSent < minSent+2 {
		maxSent = minSent + 2
	}
	td := articlegen.TemplateData{
		Items:      buildItems(words),
		Count:      count,
		MinSent:    minSent,
		MaxSent:    maxSent,
		Style:      opt.Style,
		JSONFormat: g.promptTemplate.JSONFormat,
	}

	var lastErr error
	var best *articlegen.Article // 覆盖最好(缺词最少)的一篇, 用于兜底降级
	bestMissing := -1
	attempts := opt.RetryCount + 1
	for i := 0; i < attempts; i++ {
		prompt, err := renderPrompt(g.promptTemplate.Template, td)
		if err != nil {
			return nil, errors.ErrorRequestParamError("生成提示词失败").WithCause(err)
		}

		var answer string
		if opt.Model != "" {
			answer, err = g.llm.Chat(ctx, prompt, llm.WithModel(opt.Model))
		} else {
			answer, err = g.llm.Chat(ctx, prompt)
		}
		if err != nil {
			return nil, errors.ErrorRequestParamError("AI 文章生成调用失败").WithCause(err)
		}

		art, perr := parseArticle(answer)
		if perr != nil {
			lastErr = perr
			td.Missing = "上次返回的不是有效 JSON"
			continue
		}
		normalizeArticle(art) // 清掉模型可能误加的 markdown 标记, 再校验(surface 对齐纯文本)

		// 结构问题(标题/句子缺失)是硬错误: 没有可用文章, 重试; 用尽才报错.
		if se := structuralIssue(art); se != "" {
			lastErr = fmt.Errorf("文章结构不完整: %s", se)
			td.Missing = se
			continue
		}

		// 覆盖度(软): 记录"缺词最少"的一篇作为兜底.
		missing := missingWords(art, words)
		if best == nil || len(missing) < bestMissing {
			best = art
			bestMissing = len(missing)
		}

		// 还有目标词没自然用上/对不上 -> 提示重写以尽量提高覆盖.
		if len(missing) > 0 {
			lastErr = fmt.Errorf("仍有词未用上: %s", strings.Join(missing, ", "))
			td.Missing = "以下词没被自然用进故事, 或它的形态在正文里找不到, 请重写整篇并务必自然用上它们: " + strings.Join(missing, ", ")
			continue
		}

		// 全覆盖且结构完整即可返回(不再强制"故事/过渡句", 以行文流畅为底线).
		return art, nil
	}

	// 重试用尽: 返回覆盖最好的一篇(优雅降级). 个别生僻词(如医学术语)没能自然融入时,
	// 前端只是不高亮该词, 也好过整篇生成失败.
	if best != nil {
		return best, nil
	}
	return nil, errors.ErrorRequestParamError("AI 文章生成失败").WithCause(lastErr)
}

func buildItems(words []articlegen.InputWord) []articlegen.ItemData {
	items := make([]articlegen.ItemData, 0, len(words))
	for i, w := range words {
		ts := "单词"
		if w.Type == types.WordTypePhrase {
			ts = "短语"
		}
		items = append(items, articlegen.ItemData{
			Index:   i + 1,
			Word:    w.Word,
			TypeStr: ts,
			Meaning: w.Meaning,
			Forms:   w.Forms,
		})
	}
	return items
}

func renderPrompt(tmplStr string, td articlegen.TemplateData) (string, error) {
	tmpl, err := template.New("article").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, td); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// parseArticle 清理 markdown 围栏 + 解析 JSON, 失败则用首尾大括号子串兜底.
func parseArticle(answer string) (*articlegen.Article, error) {
	cleaned := strings.TrimSpace(utils.TrimMarkDownJsonTag(answer))
	var art articlegen.Article
	if err := json.Unmarshal([]byte(cleaned), &art); err != nil {
		start := strings.Index(cleaned, "{")
		end := strings.LastIndex(cleaned, "}")
		if start < 0 || end <= start {
			return nil, fmt.Errorf("AI 返回不是有效 JSON: %s", truncateForLog(cleaned))
		}
		if err2 := json.Unmarshal([]byte(cleaned[start:end+1]), &art); err2 != nil {
			return nil, fmt.Errorf("AI 返回 JSON 解析失败: %w (原文: %s)", err2, truncateForLog(cleaned))
		}
	}
	return &art, nil
}

// structuralIssue 硬校验: 仅在"完全没有可用正文"时报错(标题已在 normalizeArticle 兜底, 空句子已被丢弃).
func structuralIssue(art *articlegen.Article) string {
	if len(art.Sentences) == 0 {
		return "正文为空"
	}
	return ""
}

// missingWords 软校验: 返回"没出现在 used_words 或 surface 命中不了正文"的输入词条原文列表.
// 用于纠错重问与"缺词最少"兜底; 不再作为硬失败(个别生僻词融不进时优雅降级).
func missingWords(art *articlegen.Article, inputs []articlegen.InputWord) []string {
	used := make(map[string]articlegen.UsedWord, len(art.UsedWords))
	for _, u := range art.UsedWords {
		used[wordKey(u.Word, u.Type)] = u
	}
	// 所有目标词文本(小写), 用于识别"surface 实为另一个目标词"的错配.
	inputSet := make(map[string]struct{}, len(inputs))
	for _, in := range inputs {
		inputSet[strings.ToLower(strings.TrimSpace(in.Word))] = struct{}{}
	}
	var missing []string
	for _, in := range inputs {
		u, ok := used[wordKey(in.Word, in.Type)]
		if !ok || len(u.Surfaces) == 0 {
			missing = append(missing, in.Word)
			continue
		}
		selfLower := strings.ToLower(strings.TrimSpace(in.Word))
		hit := false
		for _, sf := range u.Surfaces {
			sl := strings.ToLower(strings.TrimSpace(sf))
			// surface 是"别的目标词"时不算本词命中(AI 把 X 的形态错标给了 W)
			if sl != selfLower {
				if _, isOther := inputSet[sl]; isOther {
					continue
				}
			}
			if surfaceInSentences(sf, art.Sentences) {
				hit = true
				break
			}
		}
		if !hit {
			missing = append(missing, in.Word)
		}
	}
	return missing
}

func wordKey(word string, t int) string {
	return strings.ToLower(strings.TrimSpace(word)) + "|" + strconv.Itoa(t)
}

var whitespaceRe = regexp.MustCompile(`\s+`)

// surfaceInSentences 大小写不敏感 + 词边界地判断 surface 是否为某句英文的子串.
// 短语内部空白允许多空格. 编译失败时回退为大小写不敏感包含.
func surfaceInSentences(surface string, sents []articlegen.Sentence) bool {
	surface = strings.TrimSpace(surface)
	if surface == "" {
		return false
	}
	escaped := regexp.QuoteMeta(surface)
	escaped = whitespaceRe.ReplaceAllString(escaped, `\s+`)
	re, err := regexp.Compile(`(?i)\b` + escaped + `\b`)
	if err != nil {
		low := strings.ToLower(surface)
		for _, s := range sents {
			if strings.Contains(strings.ToLower(s.En), low) {
				return true
			}
		}
		return false
	}
	for _, s := range sents {
		if re.MatchString(s.En) {
			return true
		}
	}
	return false
}

// normalizeArticle 清掉模型可能误加的 markdown/符号标记; 丢弃空句子; 标题缺失时从首句兜底.
func normalizeArticle(art *articlegen.Article) {
	art.TitleEn = cleanInlineMarkdown(art.TitleEn)
	art.TitleZh = cleanInlineMarkdown(art.TitleZh)

	kept := make([]articlegen.Sentence, 0, len(art.Sentences))
	for _, s := range art.Sentences {
		en := cleanInlineMarkdown(s.En)
		zh := cleanInlineMarkdown(s.Zh)
		if en == "" || zh == "" {
			continue // 丢掉残缺句, 而不是整篇失败
		}
		kept = append(kept, articlegen.Sentence{En: en, Zh: zh})
	}
	art.Sentences = kept

	for i := range art.UsedWords {
		for j := range art.UsedWords[i].Surfaces {
			art.UsedWords[i].Surfaces[j] = cleanInlineMarkdown(art.UsedWords[i].Surfaces[j])
		}
	}

	// 标题兜底: 小模型偶尔不写标题, 有正文就从首句兜一个, 绝不因标题缺失而整篇失败.
	if strings.TrimSpace(art.TitleEn) == "" {
		art.TitleEn = deriveTitleEn(art.Sentences)
	}
	if strings.TrimSpace(art.TitleZh) == "" {
		art.TitleZh = deriveTitleZh(art.Sentences)
	}
}

func deriveTitleEn(sents []articlegen.Sentence) string {
	for _, s := range sents {
		words := strings.Fields(strings.TrimSpace(s.En))
		if len(words) == 0 {
			continue
		}
		if len(words) > 6 {
			words = words[:6]
		}
		t := strings.TrimRight(strings.Join(words, " "), ".,!?;:\"'")
		if t != "" {
			return t
		}
	}
	return "A Short Passage"
}

func deriveTitleZh(sents []articlegen.Sentence) string {
	for _, s := range sents {
		z := strings.TrimSpace(s.Zh)
		if z == "" {
			continue
		}
		r := []rune(z)
		if len(r) > 14 {
			r = r[:14]
		}
		return strings.TrimRight(string(r), "。，!?；：,.")
	}
	return "小短文"
}

// cleanInlineMarkdown 去掉星号(加粗/斜体)与反引号(代码)标记. 英语学习文本里不会出现合法的 * 或 `.
func cleanInlineMarkdown(s string) string {
	s = strings.ReplaceAll(s, "*", "")
	s = strings.ReplaceAll(s, "`", "")
	return strings.TrimSpace(s)
}

func truncateForLog(s string) string {
	const maxLen = 200
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
