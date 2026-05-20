package tts

// Options 定义了TTS的选项
type Options struct {
	Model      string // 模型名称
	Voice      string // 发音人/音色
	Accent     string // 口音，如british、american等; Voice 优先级高于 Accent
	Format     string // 音频格式，如WAV、PCM等
	SampleRate int    // 采样率，如16000
	Volume     int    // 音量大小，0-100
	SpeechRate int    // 语速，-500到500
	PitchRate  int    // 音调，-500到500
}

// Option 是用于设置Options的函数类型
type Option func(*Options)

// WithModel 设置模型
func WithModel(model string) Option {
	return func(o *Options) {
		o.Model = model
	}
}

// WithVoice 设置发音人/音色
func WithVoice(voice string) Option {
	return func(o *Options) {
		o.Voice = voice
	}
}

// WithAccent 设置口音 (如uk、us等) | 注意：Voice 优先级高于 Accent
func WithAccent(accent string) Option {
	return func(o *Options) {
		o.Accent = accent
	}
}

// WithFormat 设置音频格式
func WithFormat(format string) Option {
	return func(o *Options) {
		o.Format = format
	}
}

// WithSampleRate 设置采样率
func WithSampleRate(sampleRate int) Option {
	return func(o *Options) {
		o.SampleRate = sampleRate
	}
}

// WithVolume 设置音量大小 (0-100)
func WithVolume(volume int) Option {
	return func(o *Options) {
		o.Volume = volume
	}
}

// WithSpeechRate 设置语速 (-500到500)
func WithSpeechRate(speechRate int) Option {
	return func(o *Options) {
		o.SpeechRate = speechRate
	}
}

// WithPitchRate 设置音调 (-500到500)
func WithPitchRate(pitchRate int) Option {
	return func(o *Options) {
		o.PitchRate = pitchRate
	}
}
