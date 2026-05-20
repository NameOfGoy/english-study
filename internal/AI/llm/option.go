package llm

// Option 定义了LLM的选项
type Option struct {
	Model string
}

// Options 是用于设置option的函数类型
type Options func(*Option)

// WithModel 设置模型
func WithModel(model string) Options {
	return func(o *Option) {
		o.Model = model
	}
}