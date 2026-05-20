package wordtranslation

import (
    "context"
)

// WordInfo 是单词在 stardict 缺失时由 AI 生成的兜底信息
// 字段对应 stardict 主要字段，用于走通现有词典管线
type WordInfo struct {
    Valid       bool   // 是否合法英语单词；false 时其他字段无效
    Phonetic    string // 国际音标，例如 "/wɜːd/"
    Translation string // 多行格式 "vt. xxx\nn. xxx" 与 stardict 一致
    Definition  string // 英文释义
    Exchange    string // stardict exchange 格式 "p:did,d:done,..."（可空）
    Reason      string // 当 Valid=false 时的原因
}

// WordTranslation 是生成词语中文翻译的接口
type WordTranslation interface {
    // Generate 根据词语生成中文翻译
    Generate(ctx context.Context, word string, opts ...OptionFunc) (string, error)

    // GenerateWordInfo 在 stardict 缺失时，让 AI 生成完整的"伪 stardict"条目
    GenerateWordInfo(ctx context.Context, word string) (*WordInfo, error)
}
