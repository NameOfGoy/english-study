package articlegen

// Option 单次生成的可选参数(可覆盖默认 Model/RetryCount, 并附加风格倾向).
type Option struct {
	Model      string
	RetryCount int
	Style      string // 场景风格倾向, 可空
}

type OptionFunc func(*Option)

func WithModel(model string) OptionFunc { return func(o *Option) { o.Model = model } }

func WithRetryCount(n int) OptionFunc { return func(o *Option) { o.RetryCount = n } }

func WithStyle(s string) OptionFunc { return func(o *Option) { o.Style = s } }
