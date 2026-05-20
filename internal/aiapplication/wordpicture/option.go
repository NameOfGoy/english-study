package wordpicture

// Option 定义了生成图片的选项
type Option struct {
	Pos            string         // 词性
	Translation    string         // 释义
	Width          int            // 图片宽度
	Height         int            // 图片高度
	PromptTemplate PromptTemplate // 提示模板
}

// OptionFunc 是用于设置Option的函数类型
type OptionFunc func(*Option)

// WithPos 设置词性
func WithPos(pos string) OptionFunc {
	return func(o *Option) {
		o.Pos = pos
	}
}

// WithTranslation 设置释义
func WithTranslation(translation string) OptionFunc {
	return func(o *Option) {
		o.Translation = translation
	}
}

// WithWidth 设置图片宽度
func WithWidth(width int) OptionFunc {
	return func(o *Option) {
		o.Width = width
	}
}

// WithHeight 设置图片高度
func WithHeight(height int) OptionFunc {
	return func(o *Option) {
		o.Height = height
	}
}

// WithPromptTemplate 设置自定义的提示模板
func WithPromptTemplate(promptTemplate string) OptionFunc {
	return func(o *Option) {
		o.PromptTemplate = PromptTemplate{
			Template: promptTemplate,
		}
	}
}
