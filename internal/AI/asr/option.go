package asr

// Option 定义了ASR的选项
type option struct {
	Model string
}

// Option 是用于设置option的函数类型
type Option func(*option)

// WithModel 设置模型
func WithModel(model string) Option {
	return func(o *option) {
		o.Model = model
	}
}