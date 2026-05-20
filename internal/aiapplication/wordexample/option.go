package wordexample

type Option struct {
	Pos         string
	Count       int    // 例句数量
	Translation string // 指定的单词中文释义
}

type OptionFunc func(*Option)

func WithCount(count int) OptionFunc {
	return func(o *Option) {
		o.Count = count
	}
}

func WithTranslation(translation string) OptionFunc {
	return func(o *Option) {
		o.Translation = translation
	}
}

func WithPos(pos string) OptionFunc {
	return func(o *Option) {
		o.Pos = pos
	}
}
