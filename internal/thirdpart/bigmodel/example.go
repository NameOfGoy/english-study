package bigmodel

import (
	"context"
	"fmt"
	"log"
	"time"
)

// ExampleUsage 展示SDK的基本使用方法
func ExampleUsage() {
	// 1. 创建SDK实例
	apiKey := "your-api-key-here"
	sdk := New(apiKey)

	// 2. 验证API密钥
	if err := sdk.ValidateAPIKey(); err != nil {
		log.Fatalf("Invalid API key: %v", err)
	}

	ctx := context.Background()

	// 3. 简单对话示例
	ExampleSimpleChat(ctx, sdk)

	// 4. 流式对话示例
	ExampleStreamChat(ctx, sdk)

	// 5. 图像生成示例
	ExampleImageGeneration(ctx, sdk)

	// 6. TTS示例
	ExampleTTS(ctx, sdk)
}

// ExampleSimpleChat 简单对话示例
func ExampleSimpleChat(ctx context.Context, sdk *SDK) {
	fmt.Println("=== 简单对话示例 ===")

	// 使用便捷方法进行简单对话
	response, err := sdk.SimpleChat(ctx, ModelGLM45Flash, "你好，请介绍一下自己")
	if err != nil {
		log.Printf("Simple chat error: %v", err)
		return
	}

	fmt.Printf("AI回复: %s\n\n", response)

	// 使用带系统提示的对话
	systemMsg := "你是一个专业的英语老师，请用简洁的语言回答问题。"
	userMsg := "请解释一下什么是现在完成时？"

	response, err = sdk.ChatWithSystem(ctx, ModelGLM45Flash, systemMsg, userMsg)
	if err != nil {
		log.Printf("Chat with system error: %v", err)
		return
	}

	fmt.Printf("英语老师回复: %s\n\n", response)
}

// ExampleStreamChat 流式对话示例
func ExampleStreamChat(ctx context.Context, sdk *SDK) {
	fmt.Println("=== 流式对话示例 ===")

	// 定义流式回调函数
	callback := func(response ChatCompletionStreamResponse) error {
		if len(response.Choices) > 0 {
			content := response.Choices[0].Delta.Content
			if content != "" {
				fmt.Print(content)
			}
		}
		return nil
	}

	fmt.Print("AI流式回复: ")
	err := sdk.StreamChat(ctx, ModelGLM45Flash, "请写一首关于春天的短诗", callback)
	if err != nil {
		log.Printf("\nStream chat error: %v", err)
		return
	}
	fmt.Println()
}

// ExampleImageGeneration 图像生成示例
func ExampleImageGeneration(ctx context.Context, sdk *SDK) {
	fmt.Println("=== 图像生成示例 ===")

	// 生成简单图像
	prompt := "一只可爱的小猫在花园里玩耍，卡通风格，高质量"
	imageURL, err := sdk.GenerateImage(ctx, ModelCogView3, prompt)
	if err != nil {
		log.Printf("Image generation error: %v", err)
		return
	}

	fmt.Printf("生成的图像URL: %s\n", imageURL)

	// 生成指定尺寸的图像
	imageURL, err = sdk.GenerateImageWithSize(ctx, ModelCogView3, prompt, ImageSize1024)
	if err != nil {
		log.Printf("Image generation with size error: %v", err)
		return
	}

	fmt.Printf("1024x1024图像URL: %s\n", imageURL)

	// 生成多张图像
	imageURLs, err := sdk.GenerateMultipleImages(ctx, ModelCogView3, prompt, 2)
	if err != nil {
		log.Printf("Multiple images generation error: %v", err)
		return
	}

	fmt.Printf("生成了%d张图像:\n", len(imageURLs))
	for i, url := range imageURLs {
		fmt.Printf("  图像%d: %s\n", i+1, url)
	}

	// 生成高清图像
	hdImageURL, err := sdk.GenerateHDImage(ctx, ModelCogView3, prompt)
	if err != nil {
		log.Printf("HD image generation error: %v", err)
		return
	}

	fmt.Printf("高清图像URL: %s\n\n", hdImageURL)
}

// ExampleAdvancedUsage 高级用法示例
func ExampleAdvancedUsage() {
	fmt.Println("=== 高级用法示例 ===")

	// 1. 自定义配置
	apiKey := "your-api-key-here"
	customBaseURL := "https://custom-api.example.com/api/paas/v4"
	customTimeout := 60 * time.Second

	sdk := NewWithConfig(apiKey, customBaseURL, customTimeout)

	// 2. 使用原始API进行更精细的控制
	ctx := context.Background()

	// 自定义对话请求
	chatReq := ChatCompletionRequest{
		Model: ModelGLM45Flash,
		Messages: []Message{
			{
				Role:    "system",
				Content: "你是一个专业的代码审查员",
			},
			{
				Role:    "user",
				Content: "请审查这段Python代码的质量",
			},
		},
		Temperature: float64Ptr(0.7),
		MaxTokens:   intPtr(1000),
		TopP:        float64Ptr(0.9),
	}

	chatResp, err := sdk.Chat().CreateChatCompletion(ctx, chatReq)
	if err != nil {
		log.Printf("Advanced chat error: %v", err)
		return
	}

	fmt.Printf("代码审查结果: %s\n", GetChatResponse(chatResp))
	fmt.Printf("使用的tokens: %d\n", chatResp.Usage.TotalTokens)

	// 自定义图像生成请求
	imageReq := ImageGenerationRequest{
		Model:   ModelCogView3,
		Prompt:  "未来城市的科幻场景，赛博朋克风格",
		Size:    ImageSize1024x1792,
		N:       intPtr(1),
		Quality: ImageQualityHD,
	}

	imageResp, err := sdk.Image().CreateImage(ctx, imageReq)
	if err != nil {
		log.Printf("Advanced image generation error: %v", err)
		return
	}

	fmt.Printf("生成的科幻图像: %s\n", GetFirstImageURL(imageResp))
}

// ExampleErrorHandling 错误处理示例
func ExampleErrorHandling() {
	fmt.Println("=== 错误处理示例 ===")

	sdk := New("invalid-api-key")
	ctx := context.Background()

	_, err := sdk.SimpleChat(ctx, ModelGLM45Flash, "测试消息")
	if err != nil {
		// 检查是否为API错误
		if IsAPIError(err) {
			apiErr := GetAPIError(err)
			fmt.Printf("API错误: 状态码=%d, 消息=%s, 类型=%s\n",
				apiErr.Code, apiErr.Message, apiErr.Type)

			// 检查具体错误类型
			if IsAuthenticationError(err) {
				fmt.Println("这是认证错误，请检查API密钥")
			} else if IsRateLimitError(err) {
				fmt.Println("请求频率超限，请稍后重试")
			} else if IsRetryableError(err) {
				fmt.Println("这是可重试的错误")
			}
		} else {
			fmt.Printf("其他错误: %v\n", err)
		}
	}
}

// ExampleTTS TTS示例
func ExampleTTS(ctx context.Context, sdk *SDK) {
	fmt.Println("=== TTS示例 ===")

	// 1. 简单TTS
	fmt.Println("1. 简单TTS:")
	text := "你好，欢迎使用智谱AI的文本转语音功能！"
	response, err := sdk.TTS().CreateSimpleTTS(ctx, text)
	if err != nil {
		log.Printf("Simple TTS error: %v", err)
		return
	}

	// 保存音频文件
	err = SaveAudioToFile(response, "simple_tts.wav")
	if err != nil {
		log.Printf("Save audio file error: %v", err)
		return
	}
	fmt.Printf("简单TTS音频已保存到: %s\n", response.Filename)
	fmt.Printf("音频数据大小: %d bytes\n\n", len(response.Data))

	// 2. 带语音配置的TTS
	fmt.Println("2. 带语音配置的TTS:")
	text2 := "你好，这是一个语音测试，使用了不同的音色和语速。"
	response2, err := sdk.TTS().CreateTTSWithVoice(ctx, text2, TTSVoiceChuichui)
	if err != nil {
		log.Printf("TTS with voice error: %v", err)
		return
	}

	err = SaveAudioToFile(response2, "voice_config_tts.wav")
	if err != nil {
		log.Printf("Save audio file error: %v", err)
		return
	}
	fmt.Printf("带语音配置的TTS音频已保存到: %s\n", response2.Filename)
	fmt.Printf("使用音色: %s\n\n", TTSVoiceChuichui)

	// 3. 完整配置的TTS
	fmt.Println("3. 完整配置的TTS:")
	text3 := "这是一个完整配置的语音合成示例，包含了语速、音量和音频格式设置。"
	speed := 1.2
	volume := 0.8
	format := TTSResponseFormatMP3
	response3, err := sdk.TTS().CreateTTSWithConfig(ctx, text3, TTSVoiceXiaochen, &speed, &volume, &format)
	if err != nil {
		log.Printf("TTS with config error: %v", err)
		return
	}

	err = SaveAudioToFile(response3, "config_tts.mp3")
	if err != nil {
		log.Printf("Save audio file error: %v", err)
		return
	}
	fmt.Printf("完整配置的TTS音频已保存到: %s\n", response3.Filename)
	fmt.Printf("使用音色: %s, 语速: %.1f, 音量: %.1f, 格式: %s\n\n", TTSVoiceXiaochen, speed, volume, format)

	// 4. 直接使用TTSRequest
	fmt.Println("4. 直接使用TTSRequest:")
	request := TTSRequest{
		Model:  "cogtts",
		Input:  "这是一个直接使用TTSRequest的示例，展示了最灵活的调用方式。",
		Voice:  string(TTSVoiceJam),
		Speed:  &[]float64{0.8}[0],
		Volume: &[]float64{1.5}[0],
	}

	response4, err := sdk.TTS().CreateTTS(ctx, request)
	if err != nil {
		log.Printf("Direct TTS request error: %v", err)
		return
	}

	err = SaveAudioToFile(response4, "direct_tts.wav")
	if err != nil {
		log.Printf("Save audio file error: %v", err)
		return
	}
	fmt.Printf("直接请求的TTS音频已保存到: %s\n", response4.Filename)
	fmt.Printf("使用音色: %s, 语速: %.1f, 音量: %.1f\n\n", TTSVoiceJam, *request.Speed, *request.Volume)

	// 5. 保存到指定文件
	fmt.Println("5. 保存到指定文件:")
	text5 := "这是保存到指定文件的示例。"
	response5, err := sdk.TTS().CreateTTSWithVoice(ctx, text5, TTSVoiceKazi)
	if err != nil {
		log.Printf("TTS error: %v", err)
		return
	}

	// 使用SaveToFile方法保存到指定文件
	err = sdk.TTS().SaveToFile(response5, "custom_filename.wav")
	if err != nil {
		log.Printf("Save to file error: %v", err)
		return
	}
	fmt.Printf("TTS音频已保存到指定文件: custom_filename.wav\n")
	fmt.Printf("使用音色: %s\n\n", TTSVoiceKazi)

	fmt.Println("TTS示例完成！")
}
