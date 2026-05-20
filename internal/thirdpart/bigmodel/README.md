# 智谱AI Go SDK

智谱AI官方Go语言SDK，支持对话补全和图像生成功能。

## 特性

- 🚀 **简单易用**: 提供简洁的API接口
- 🔐 **安全认证**: 支持Bearer Token认证
- 💬 **对话补全**: 支持流式和非流式对话
- 🎨 **图像生成**: 支持多种尺寸和质量的图像生成
- 🔊 **文本转语音**: 支持cogtts模型的TTS功能
- 🛡️ **错误处理**: 完善的错误处理机制
- 📦 **框架化**: 模块化设计，易于扩展

## 安装

```bash
go get github.com/your-org/bigmodel-sdk
```

## 快速开始

### 1. 初始化SDK

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/your-org/bigmodel-sdk"
)

func main() {
    // 创建SDK实例
    sdk := bigmodel.New("your-api-key")
    
    // 验证API密钥
    if err := sdk.ValidateAPIKey(); err != nil {
        log.Fatal("Invalid API key:", err)
    }
    
    ctx := context.Background()
    
    // 简单对话
    response, err := sdk.SimpleChat(ctx, bigmodel.ModelGLM4, "你好")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("AI回复:", response)
}
```

### 2. 对话补全

#### 简单对话

```go
// 基础对话
response, err := sdk.SimpleChat(ctx, bigmodel.ModelGLM4, "请介绍一下Go语言")
if err != nil {
    log.Fatal(err)
}
fmt.Println(response)

// 带系统提示的对话
systemMsg := "你是一个专业的程序员助手"
userMsg := "如何优化Go程序的性能？"
response, err = sdk.ChatWithSystem(ctx, bigmodel.ModelGLM4, systemMsg, userMsg)
if err != nil {
    log.Fatal(err)
}
fmt.Println(response)
```

#### 流式对话

```go
// 定义流式回调
callback := func(response bigmodel.ChatCompletionStreamResponse) error {
    if len(response.Choices) > 0 {
        content := response.Choices[0].Delta.Content
        if content != "" {
            fmt.Print(content)
        }
    }
    return nil
}

// 开始流式对话
err := sdk.StreamChat(ctx, bigmodel.ModelGLM4, "写一首诗", callback)
if err != nil {
    log.Fatal(err)
}
```

#### 高级对话

```go
// 自定义请求参数
req := bigmodel.ChatCompletionRequest{
    Model: bigmodel.ModelGLM4,
    Messages: []bigmodel.Message{
        {
            Role:    "system",
            Content: "你是一个专业的翻译助手",
        },
        {
            Role:    "user",
            Content: "请将以下文本翻译成英文：你好世界",
        },
    },
    Temperature: bigmodel.Float64Ptr(0.7),
    MaxTokens:   bigmodel.IntPtr(1000),
    TopP:        bigmodel.Float64Ptr(0.9),
}

resp, err := sdk.Chat().CreateChatCompletion(ctx, req)
if err != nil {
    log.Fatal(err)
}

fmt.Println("翻译结果:", bigmodel.GetChatResponse(resp))
fmt.Printf("使用tokens: %d\n", resp.Usage.TotalTokens)
```

### 3. 图像生成

#### 简单图像生成

```go
// 生成基础图像
imageURL, err := sdk.GenerateImage(ctx, bigmodel.ModelCogView3, "一只可爱的小猫")
if err != nil {
    log.Fatal(err)
}
fmt.Println("图像URL:", imageURL)

// 生成指定尺寸的图像
imageURL, err = sdk.GenerateImageWithSize(ctx, bigmodel.ModelCogView3, "美丽的风景", bigmodel.ImageSize1024)
if err != nil {
    log.Fatal(err)
}
fmt.Println("1024x1024图像:", imageURL)

// 生成多张图像
imageURLs, err := sdk.GenerateMultipleImages(ctx, bigmodel.ModelCogView3, "抽象艺术", 3)
if err != nil {
    log.Fatal(err)
}
for i, url := range imageURLs {
    fmt.Printf("图像%d: %s\n", i+1, url)
}

// 生成高清图像
hdImageURL, err := sdk.GenerateHDImage(ctx, bigmodel.ModelCogView3, "科幻城市")
if err != nil {
    log.Fatal(err)
}
fmt.Println("高清图像:", hdImageURL)
```

#### 高级图像生成

```go
// 自定义图像生成请求
req := bigmodel.ImageGenerationRequest{
    Model:   bigmodel.ModelCogView3,
    Prompt:  "未来城市，赛博朋克风格，高质量渲染",
    Size:    bigmodel.ImageSize1024x1792,
    N:       bigmodel.IntPtr(2),
    Quality: bigmodel.ImageQualityHD,
}

resp, err := sdk.Image().CreateImage(ctx, req)
if err != nil {
    log.Fatal(err)
}

for i, url := range bigmodel.GetImageURLs(resp) {
    fmt.Printf("生成的图像%d: %s\n", i+1, url)
}
```

### 4. 文本转语音 (TTS)

#### 简单TTS

```go
// 基础文本转语音
response, err := sdk.TTS().CreateSimpleTTS(ctx, "你好，欢迎使用智谱AI语音合成服务")
if err != nil {
    log.Fatal(err)
}

// 保存音频文件
err = sdk.TTS().SaveAudioToFile(response, "output.wav")
if err != nil {
    log.Fatal(err)
}
fmt.Println("音频文件已保存为 output.wav")
```

#### 带语音配置的TTS

```go
// 使用指定音色和语速
response, err := sdk.TTS().CreateTTSWithVoice(ctx, "这是一段快乐的语音", "chuichui")
if err != nil {
    log.Fatal(err)
}

// 保存音频文件
err = sdk.TTS().SaveAudioToFile(response, "voice_output.wav")
if err != nil {
    log.Fatal(err)
}
```

#### 完整配置的TTS

```go
// 使用完整配置进行TTS
req := bigmodel.TTSRequest{
    Model:  "cogtts",
    Input:  "这是一段需要转换为语音的文本内容",
    Voice:  "tongtong",
    Speed:  1.2,
    Volume: 0.8,
    ResponseFormat: "wav",
    WatermarkEnabled: true,
}

response, err := sdk.TTS().CreateTTSWithConfig(ctx, req)
if err != nil {
    log.Fatal(err)
}

// 保存音频文件
err = sdk.TTS().SaveAudioToFile(response, "configured_output.wav")
if err != nil {
    log.Fatal(err)
}
```

#### 直接使用TTSRequest

```go
// 直接使用TTSRequest进行TTS
req := bigmodel.TTSRequest{
    Model:  "cogtts",
    Input:  "欢迎使用智谱AI的文本转语音服务，我们提供高质量的语音合成功能",
    Voice:  "jam",
    Speed:  1.1,
    Volume: 0.9,
    ResponseFormat: "wav",
    WatermarkEnabled: false,
}

response, err := sdk.TTS().CreateTTS(ctx, req)
if err != nil {
    log.Fatal(err)
}

// 保存音频文件
err = sdk.TTS().SaveAudioToFile(response, "direct_request.wav")
if err != nil {
    log.Fatal(err)
}

#### 保存到指定文件

```go
// 保存到指定文件名
req := bigmodel.TTSRequest{
    Model: "cogtts",
    Input: "这是保存到指定文件的示例",
    Voice: "xiaochen",
}

response, err := sdk.TTS().CreateTTS(ctx, req)
if err != nil {
    log.Fatal(err)
}

// 使用SaveToFile方法保存到指定文件
err = sdk.TTS().SaveToFile(response, "custom_filename.wav")
if err != nil {
    log.Fatal(err)
}
```

## 配置选项

### 自定义配置

```go
// 使用自定义配置创建SDK
sdk := bigmodel.NewWithConfig(
    "your-api-key",
    "https://custom-api.example.com/api/paas/v4", // 自定义API地址
    60*time.Second, // 自定义超时时间
)

// 运行时修改配置
sdk.SetTimeout(30 * time.Second)
sdk.SetAPIKey("new-api-key")
```

## 支持的模型

### 对话模型

- `bigmodel.ModelGLM4` - GLM-4 模型
- `bigmodel.ModelGLM4V` - GLM-4V 多模态模型
- `bigmodel.ModelGLM3Turbo` - GLM-3-Turbo 模型
- `bigmodel.ModelGLM4Flash` - GLM-4-Flash 快速模型

### 图像生成模型

- `bigmodel.ModelCogView3` - CogView-3 图像生成模型

### TTS模型

- `cogtts` - CogTTS 文本转语音模型

### TTS参数说明

#### 支持的音色 (Voice)
- `tongtong` - 默认音色
- `chuichui` - 吹吹音色
- `xiaochen` - 小陈音色
- `jam` - Jam音色
- `kazi` - 卡兹音色
- `douji` - 豆几音色
- `luodo` - 罗朵音色

#### 参数范围
- **语速 (Speed)**: 0.5 - 2.0，默认 1.0
- **音量 (Volume)**: 0.1 - 10.0，默认 1.0
- **响应格式 (ResponseFormat)**: wav（默认）、mp3
- **水印 (WatermarkEnabled)**: true（默认）、false
- **文本长度**: 最大 4096 字符

## 图像尺寸选项

- `bigmodel.ImageSize256` - 256x256
- `bigmodel.ImageSize512` - 512x512
- `bigmodel.ImageSize1024` - 1024x1024
- `bigmodel.ImageSize1024x1792` - 1024x1792
- `bigmodel.ImageSize1792x1024` - 1792x1024

## 错误处理

```go
_, err := sdk.SimpleChat(ctx, bigmodel.ModelGLM4, "测试")
if err != nil {
    // 检查是否为API错误
    if bigmodel.IsAPIError(err) {
        apiErr := bigmodel.GetAPIError(err)
        fmt.Printf("API错误: %d - %s\n", apiErr.Code, apiErr.Message)
        
        // 检查具体错误类型
        if bigmodel.IsAuthenticationError(err) {
            fmt.Println("认证失败，请检查API密钥")
        } else if bigmodel.IsRateLimitError(err) {
            fmt.Println("请求频率超限")
        } else if bigmodel.IsRetryableError(err) {
            fmt.Println("可重试的错误")
        }
    } else {
        fmt.Printf("其他错误: %v\n", err)
    }
}
```

## API参考

### SDK主要方法

| 方法 | 描述 |
|------|------|
| `New(apiKey)` | 创建SDK实例 |
| `NewWithConfig(apiKey, baseURL, timeout)` | 使用自定义配置创建SDK |
| `ValidateAPIKey()` | 验证API密钥 |
| `SetTimeout(duration)` | 设置超时时间 |
| `SetAPIKey(key)` | 设置API密钥 |

### 对话相关方法

| 方法 | 描述 |
|------|------|
| `SimpleChat(ctx, model, message)` | 简单对话 |
| `ChatWithSystem(ctx, model, system, user)` | 带系统提示的对话 |
| `StreamChat(ctx, model, message, callback)` | 流式对话 |
| `Chat().CreateChatCompletion(ctx, req)` | 自定义对话请求 |
| `Chat().CreateChatCompletionStream(ctx, req, callback)` | 自定义流式对话 |

### 图像生成方法

| 方法 | 描述 |
|------|------|
| `GenerateImage(ctx, model, prompt)` | 生成图像 |
| `GenerateImageWithSize(ctx, model, prompt, size)` | 生成指定尺寸图像 |
| `GenerateMultipleImages(ctx, model, prompt, count)` | 生成多张图像 |
| `GenerateHDImage(ctx, model, prompt)` | 生成高清图像 |
| `Image().CreateImage(ctx, req)` | 自定义图像生成请求 |

### TTS方法

| 方法 | 描述 |
|------|------|
| `TTS().CreateSimpleTTS(ctx, text)` | 简单文本转语音 |
| `TTS().CreateTTSWithVoice(ctx, text, voice)` | 带指定音色的TTS |
| `TTS().CreateTTSWithConfig(ctx, req)` | 使用完整配置的TTS |
| `TTS().CreateTTS(ctx, req)` | 直接使用TTSRequest |
| `TTS().SaveAudioToFile(response, filename)` | 保存音频到文件 |
| `TTS().SaveToFile(response, filename)` | 保存到指定文件名 |

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request！

## 更新日志

### v1.0.0
- 初始版本发布
- 支持对话补全功能
- 支持图像生成功能
- 完善的错误处理机制