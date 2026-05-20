package alicloud

import (
	"context"
	"fmt"

	"english-study/internal/AI/tts"
	"english-study/internal/thirdpart/alibabacloud"
	"english-study/internal/utils"
)

// AliCloudTTS 阿里云语音合成实现
type AliCloudTTS struct {
	client *alibabacloud.TTSClient
}

// NewAliCloudTTS 创建阿里云TTS实例
func NewAliCloudTTS(config *alibabacloud.TTSConfig) *AliCloudTTS {
	client := alibabacloud.NewTTSClient(config)
	return &AliCloudTTS{
		client: client,
	}
}

// accentToVoiceMap 口音到发音人的映射
var accentToVoiceMap = map[string]string{
	"american": "Donna", // 美式英文女声
	"british":  "Lydia", // 英式英文女声
	"us":       "Donna", // 美式英文别名
	"uk":       "Lydia", // 英式英文别名
}

// TextToSpeech 实现TTS接口的文本转语音方法
func (a *AliCloudTTS) TextToSpeech(ctx context.Context, text string, opts ...tts.Option) ([]byte, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	// 解析选项
	opt := utils.GetOptionFromOptions[tts.Options, tts.Option](opts, func() tts.Options {
		return tts.Options{}
	})

	// 确定最终使用的发音人
	voice := a.resolveVoice(opt)

	// 构建阿里云TTS选项
	ttsOpts := &alibabacloud.TTSOptions{
		Voice:      voice,
		Format:     opt.Format,
		SampleRate: opt.SampleRate,
		Volume:     opt.Volume,
		SpeechRate: opt.SpeechRate,
		PitchRate:  opt.PitchRate,
	}

	// 调用阿里云TTS客户端
	audioData, err := a.client.TextToSpeechToBytes(ctx, text, ttsOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to convert text to speech: %w", err)
	}

	return audioData, nil
}

// resolveVoice 根据优先级规则确定最终使用的发音人
// 优先级：Voice > Accent > Model
func (a *AliCloudTTS) resolveVoice(opt tts.Options) string {
	// 1. 如果直接指定了Voice，优先使用
	if opt.Voice != "" {
		return opt.Voice
	}

	// 2. 如果指定了Accent，根据映射转换为发音人
	if opt.Accent != "" {
		if voice, exists := accentToVoiceMap[opt.Accent]; exists {
			return voice
		}
	}

	// 3. 如果设置了Model，使用Model作为Voice（向后兼容）
	if opt.Model != "" {
		return opt.Model
	}

	// 4. 默认返回空字符串，让阿里云SDK使用默认发音人
	return ""
}
