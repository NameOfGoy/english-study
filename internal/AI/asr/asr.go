package asr

import "context"

// ASR 语音识别接口
type ASR interface {
	// SpeechToText 语音转文本
	SpeechToText(ctx context.Context, audio []byte, opts ...Option) (text string, err error)
}