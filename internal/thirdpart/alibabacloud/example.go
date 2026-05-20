package alibabacloud

import (
	"context"
	"fmt"
	"log"
	"time"
)

// ExampleUsage 展示如何使用阿里云语音合成客户端
func ExampleUsage() {
	// 配置阿里云语音合成客户端
	config := &TTSConfig{
		AccessKeyID:     "your-access-key-id",     // 替换为你的AccessKey ID
		AccessKeySecret: "your-access-key-secret", // 替换为你的AccessKey Secret
		AppKey:          "your-app-key",           // 替换为你的AppKey
		Region:          "cn-shanghai",                    // 可选，默认cn-shanghai
	}

	// 创建客户端
	client := NewTTSClient(config)

	// 创建上下文，设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 示例1：使用默认选项进行语音合成
	text1 := "你好，这是一个语音合成测试。"
	audioData, err := client.TextToSpeechToBytes(ctx, text1, nil)
	if err != nil {
		log.Printf("语音合成失败: %v", err)
		return
	}
	fmt.Printf("合成成功，音频数据大小: %d 字节\n", len(audioData))

	// 示例2：使用自定义选项进行语音合成并保存到文件
	text2 := "Hello, this is a text-to-speech example with custom options."
	customOptions := &TTSOptions{
		Voice:      "xiaogang", // 男声
		Format:     "WAV",
		SampleRate: 16000,
		Volume:     80,
		SpeechRate: 100, // 稍快语速
		PitchRate:  50,  // 稍高音调
	}

	outputFile := "output.wav"
	err = client.TextToSpeechToFile(ctx, text2, customOptions, outputFile)
	if err != nil {
		log.Printf("语音合成并保存文件失败: %v", err)
		return
	}
	fmt.Printf("语音合成成功，已保存到文件: %s\n", outputFile)

	// 示例3：使用完整的TextToSpeech方法，获取详细结果
	text3 := "这是一个带字幕的语音合成示例。"
	subtitleOptions := &TTSOptions{
		Voice:          "xiaoyun",
		EnableSubtitle: true, // 开启字幕功能
	}

	result, err := client.TextToSpeech(ctx, text3, subtitleOptions, "")
	if err != nil {
		log.Printf("语音合成失败: %v", err)
		return
	}

	fmt.Printf("合成成功:\n")
	fmt.Printf("- 音频数据大小: %d 字节\n", len(result.AudioData))
	if result.Subtitle != "" {
		fmt.Printf("- 字幕信息: %s\n", result.Subtitle)
	}
}

// SimpleTextToSpeech 简单的文本转语音函数
func SimpleTextToSpeech(accessKeyID, accessKeySecret, appKey, text, outputFile string) error {
	config := &TTSConfig{
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
		AppKey:          appKey,
	}

	client := NewTTSClient(config)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	return client.TextToSpeechToFile(ctx, text, nil, outputFile)
}
