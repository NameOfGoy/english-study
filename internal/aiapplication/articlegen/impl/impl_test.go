package impl

import (
	"context"
	"english-study/internal/AI/llm"
	"english-study/internal/aiapplication/articlegen"
	"strings"
	"testing"
)

// fakeLLM 可脚本化多次回复, 用于覆盖纠错重问路径; 永不触网.
type fakeLLM struct {
	replies []string
	calls   int
}

func (f *fakeLLM) Chat(_ context.Context, _ string, _ ...llm.Options) (string, error) {
	idx := f.calls
	if idx >= len(f.replies) {
		idx = len(f.replies) - 1
	}
	f.calls++
	return f.replies[idx], nil
}

func (f *fakeLLM) StreamChat(_ context.Context, _ string, _ ...llm.Options) (<-chan string, error) {
	ch := make(chan string)
	close(ch)
	return ch, nil
}

func sampleInputs() []articlegen.InputWord {
	return []articlegen.InputWord{
		{Word: "run", Type: 1, Meaning: "跑"},
		{Word: "brave", Type: 1, Meaning: "勇敢的"},
		{Word: "give up", Type: 2, Meaning: "放弃"},
	}
}

const goodJSON = `{"title_en":"A Brave Run","title_zh":"勇敢的奔跑",` +
	`"sentences":[{"en":"He likes to run fast.","zh":"他喜欢跑得快。"},` +
	`{"en":"She was brave and did not give up.","zh":"她很勇敢，没有放弃。"}],` +
	`"used_words":[{"word":"run","type":1,"surfaces":["run"]},` +
	`{"word":"brave","type":1,"surfaces":["brave"]},` +
	`{"word":"give up","type":2,"surfaces":["give up"]}]}`

// 缺少 "give up" 的不合格回复
const missingJSON = `{"title_en":"T","title_zh":"标题",` +
	`"sentences":[{"en":"He likes to run fast.","zh":"他跑。"},{"en":"She was brave.","zh":"她勇敢。"}],` +
	`"used_words":[{"word":"run","type":1,"surfaces":["run"]},{"word":"brave","type":1,"surfaces":["brave"]}]}`

func genWith(replies ...string) (*Generator, *fakeLLM) {
	f := &fakeLLM{replies: replies}
	return NewGenerator(f, articlegen.Config{RetryCount: 1}), f
}

func TestGenerate_HappyPath(t *testing.T) {
	g, f := genWith(goodJSON)
	art, err := g.Generate(context.Background(), sampleInputs())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.calls != 1 {
		t.Fatalf("expected 1 LLM call, got %d", f.calls)
	}
	if len(art.UsedWords) != 3 || len(art.Sentences) != 2 {
		t.Fatalf("unexpected article shape: %+v", art)
	}
}

func TestGenerate_FencedJSON(t *testing.T) {
	g, _ := genWith("```json\n" + goodJSON + "\n```")
	if _, err := g.Generate(context.Background(), sampleInputs()); err != nil {
		t.Fatalf("fenced json should parse: %v", err)
	}
}

func TestGenerate_ProseWrappedJSON(t *testing.T) {
	g, _ := genWith("好的，这是结果：\n" + goodJSON + "\n以上。")
	if _, err := g.Generate(context.Background(), sampleInputs()); err != nil {
		t.Fatalf("prose-wrapped json should parse via fallback: %v", err)
	}
}

func TestGenerate_RetrySucceeds(t *testing.T) {
	g, f := genWith(missingJSON, goodJSON)
	if _, err := g.Generate(context.Background(), sampleInputs()); err != nil {
		t.Fatalf("retry should succeed: %v", err)
	}
	if f.calls != 2 {
		t.Fatalf("expected exactly 2 LLM calls, got %d", f.calls)
	}
}

// 缺词无法补齐时不再硬失败, 而是降级返回"缺词最少"的那篇(生僻词融不进也好过整篇失败).
func TestGenerate_DegradesWhenWordMissing(t *testing.T) {
	g, f := genWith(missingJSON, missingJSON)
	art, err := g.Generate(context.Background(), sampleInputs())
	if err != nil {
		t.Fatalf("缺词应降级返回最佳文章而非报错, 但收到: %v", err)
	}
	if f.calls != 2 {
		t.Fatalf("expected exactly 2 LLM calls, got %d", f.calls)
	}
	if art == nil || len(art.Sentences) == 0 {
		t.Fatal("应返回兜底文章")
	}
}

// 模型没写标题但有正文 -> 兜底从首句生成标题, 不报错.
func TestGenerate_BackfillsMissingTitle(t *testing.T) {
	const noTitleJSON = `{"title_en":"","title_zh":"",` +
		`"sentences":[{"en":"He likes to run fast.","zh":"他喜欢跑得快。"},` +
		`{"en":"She was brave and did not give up.","zh":"她很勇敢，没有放弃。"}],` +
		`"used_words":[{"word":"run","type":1,"surfaces":["run"]},` +
		`{"word":"brave","type":1,"surfaces":["brave"]},` +
		`{"word":"give up","type":2,"surfaces":["give up"]}]}`
	g, _ := genWith(noTitleJSON)
	art, err := g.Generate(context.Background(), sampleInputs())
	if err != nil {
		t.Fatalf("缺标题应兜底而非报错, 但收到: %v", err)
	}
	if strings.TrimSpace(art.TitleEn) == "" || strings.TrimSpace(art.TitleZh) == "" {
		t.Fatalf("标题应被兜底填充, 实际 en=%q zh=%q", art.TitleEn, art.TitleZh)
	}
}

// 完全没有可用正文才硬失败.
func TestGenerate_ErrorsWhenStructurallyBroken(t *testing.T) {
	const brokenJSON = `{"title_en":"","title_zh":"","sentences":[],"used_words":[]}`
	g, _ := genWith(brokenJSON, brokenJSON)
	if _, err := g.Generate(context.Background(), sampleInputs()); err == nil {
		t.Fatal("结构损坏(无标题/正文)应报错")
	}
}

// 模型把目标词用 **粗体** / *斜体* 包起来时, 后端必须清成纯文本(渲染与存库都不带星号).
const markdownJSON = `{"title_en":"A **Brave** Run","title_zh":"勇敢的**奔跑**",` +
	`"sentences":[{"en":"He likes to **run** fast.","zh":"他喜欢跑。"},` +
	`{"en":"She was *brave* and did not give up.","zh":"她很勇敢，没有放弃。"}],` +
	`"used_words":[{"word":"run","type":1,"surfaces":["run"]},` +
	`{"word":"brave","type":1,"surfaces":["brave"]},` +
	`{"word":"give up","type":2,"surfaces":["give up"]}]}`

func TestGenerate_StripsMarkdown(t *testing.T) {
	g, _ := genWith(markdownJSON)
	art, err := g.Generate(context.Background(), sampleInputs())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(art.TitleEn, "*") || strings.Contains(art.TitleZh, "*") {
		t.Fatalf("标题仍含星号: %q / %q", art.TitleEn, art.TitleZh)
	}
	for _, s := range art.Sentences {
		if strings.Contains(s.En, "*") || strings.Contains(s.Zh, "*") {
			t.Fatalf("句子仍含星号: %q", s.En)
		}
	}
}

// AI 把某词的 surface 错标成"另一个目标词"时, 该词应判为缺失(不能靠别人的词蒙混过关).
func TestMissingWords_RejectsCrossWordSurface(t *testing.T) {
	art := &articlegen.Article{
		TitleEn: "T", TitleZh: "标题",
		Sentences: []articlegen.Sentence{{En: "She was brave today.", Zh: "她今天很勇敢。"}},
		UsedWords: []articlegen.UsedWord{
			{Word: "grid", Type: 1, Surfaces: []string{"brave"}}, // 错配: grid 的 surface 是另一个目标词
			{Word: "brave", Type: 1, Surfaces: []string{"brave"}},
		},
	}
	inputs := []articlegen.InputWord{{Word: "grid", Type: 1}, {Word: "brave", Type: 1}}
	missing := missingWords(art, inputs)
	hasGrid, hasBrave := false, false
	for _, m := range missing {
		if m == "grid" {
			hasGrid = true
		}
		if m == "brave" {
			hasBrave = true
		}
	}
	if !hasGrid {
		t.Fatalf("grid 的 surface 实为另一个目标词, 应判缺失; 实际 missing=%v", missing)
	}
	if hasBrave {
		t.Fatalf("brave 应正常命中, 不该缺失; missing=%v", missing)
	}
}

func TestSurfaceBoundary(t *testing.T) {
	sents := []articlegen.Sentence{{En: "I ate a bran muffin.", Zh: "x"}}
	if surfaceInSentences("ran", sents) {
		t.Fatal("'ran' must not match inside 'bran'")
	}
	if !surfaceInSentences("bran", sents) {
		t.Fatal("'bran' should match")
	}
	if !surfaceInSentences("BRAN", sents) {
		t.Fatal("matching should be case-insensitive")
	}
	phrase := []articlegen.Sentence{{En: "He finally gave  up early.", Zh: "x"}}
	if !surfaceInSentences("gave up", phrase) {
		t.Fatal("phrase surface should match across extra whitespace")
	}
}
