package articlegen

import "context"

// InputWord 喂给 prompt 的词条 (含释义提示 + 可接受变形, 帮 AI 用对词义并辅助高亮校验).
type InputWord struct {
	Word     string // 词条原文(原型)
	Type     int    // 1-单词 2-短语
	PosLabel string // 词性缩写(单词); 短语为空
	Meaning  string // 中文释义提示(取 WordCard.Translation 首行)
	Forms    string // 可接受变形, 逗号分隔, 可空(来自 Exchange)
}

// Sentence 双语句子.
type Sentence struct {
	En string `json:"en"`
	Zh string `json:"zh"`
}

// UsedWord AI 实际用到的词条 + 其在正文中的字面形态(含变形), 供前端高亮.
type UsedWord struct {
	Word     string   `json:"word"`     // 对应 InputWord.Word(原型)
	Type     int      `json:"type"`     // 1-单词 2-短语
	Surfaces []string `json:"surfaces"` // 正文中出现的字面形式
}

// Article 生成结果(尚未持久化). 不含 pos/释义/音标 —— 由后端用本地词典补齐.
type Article struct {
	TitleEn   string     `json:"title_en"`
	TitleZh   string     `json:"title_zh"`
	Sentences []Sentence `json:"sentences"`
	UsedWords []UsedWord `json:"used_words"`
}

// Config 文章生成配置(来源 config.BigModel.Article, 由 main 转换后传入, 故本包不依赖 config).
type Config struct {
	Model      string // 生成所用模型, 空则用 llm 默认
	RetryCount int    // 校验失败纠错重问次数, 默认 1 => 总调用 <=2
}

// ArticleGenerator 由词条列表生成一篇含全部词条的双语短文.
type ArticleGenerator interface {
	Generate(ctx context.Context, words []InputWord, opts ...OptionFunc) (*Article, error)
}
