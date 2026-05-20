package tts

import "context"

// TTS 文本转语音接口
type TTS interface {
	// TextToSpeech 文本转语音
	TextToSpeech(ctx context.Context, text string, opts ...Option) ([]byte, error)
}