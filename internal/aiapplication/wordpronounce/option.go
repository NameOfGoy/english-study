package wordpronounce

// Option 语音生成选项
type Option struct {
	Accent       string // 口音
	PartOfSpeech string // 词性
	Definition   string // 释义
}

// OptionFunc 选项函数类型
type OptionFunc func(*Option)

// WithAccent 设置口音 | uk/us
func WithAccent(accent string) OptionFunc {
	return func(o *Option) {
		o.Accent = accent
	}
}

// WithPartOfSpeech 设置词性
func WithPartOfSpeech(pos string) OptionFunc {
	return func(o *Option) {
		o.PartOfSpeech = pos
	}
}

// WithDefinition 设置释义
func WithDefinition(definition string) OptionFunc {
	return func(o *Option) {
		o.Definition = definition
	}
}
