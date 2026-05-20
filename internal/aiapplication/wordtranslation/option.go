package wordtranslation

// Option 定义了生成翻译的选项
type Option struct {
    Pos string // 词性（可选，便于限定语义范围）
}

// OptionFunc 是用于设置Option的函数类型
type OptionFunc func(*Option)

// WithPos 设置词性（中文名称），帮助模型限定翻译语境
func WithPos(pos string) OptionFunc {
    return func(o *Option) {
        o.Pos = pos
    }
}