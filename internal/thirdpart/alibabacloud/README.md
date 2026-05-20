# 阿里云语音合成 Go SDK 集成

这是一个简单的阿里云语音合成（Text-to-Speech）Go SDK 集成包装器，基于阿里云官方的 `alibabacloud-nls-go-sdk`。

## 功能特性

- 简单易用的客户端接口
- 支持多种音频格式（WAV、PCM、OPUS等）
- 可自定义语音参数（发音人、语速、音调、音量等）
- 支持字幕功能
- 支持保存到文件或返回音频字节数据
- 完善的错误处理和超时控制
- 线程安全

## 安装依赖

```bash
go get github.com/aliyun/alibabacloud-nls-go-sdk
```

## 快速开始

### 1. 配置客户端

```go
package main

import (
    "context"
    "fmt"
    "time"
    "your-project/internal/thirdpart/alibabacloud"
)

func main() {
    // 配置阿里云语音合成客户端
    config := &alibabacloud.TTSConfig{
        AccessKeyID:     "your-access-key-id",     // 阿里云AccessKey ID
        AccessKeySecret: "your-access-key-secret", // 阿里云AccessKey Secret
        AppKey:          "your-app-key",           // 语音服务AppKey
        Region:          "cn-shanghai",            // 可选，默认cn-shanghai
    }

    // 创建客户端
    client := alibabacloud.NewTTSClient(config)
}
```

### 2. 基本使用

#### 获取音频字节数据

```go
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
defer cancel()

// 使用默认选项
audioData, err := client.TextToSpeechToBytes(ctx, "你好，这是一个语音合成测试。", nil)
if err != nil {
    fmt.Printf("语音合成失败: %v\n", err)
    return
}

fmt.Printf("合成成功，音频数据大小: %d 字节\n", len(audioData))
```

#### 保存音频到文件

```go
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
defer cancel()

err := client.TextToSpeechToFile(ctx, "Hello, World!", nil, "output.wav")
if err != nil {
    fmt.Printf("语音合成失败: %v\n", err)
    return
}

fmt.Println("语音合成成功，已保存到 output.wav")
```

### 3. 自定义选项

```go
// 自定义语音合成选项
options := &alibabacloud.TTSOptions{
    Voice:          "xiaogang",  // 发音人（男声）
    Format:         "WAV",       // 音频格式
    SampleRate:     16000,       // 采样率
    Volume:         80,          // 音量 (0-100)
    SpeechRate:     100,         // 语速 (-500到500)
    PitchRate:      50,          // 音调 (-500到500)
    EnableSubtitle: true,        // 开启字幕功能
}

result, err := client.TextToSpeech(ctx, "这是一个自定义选项的示例", options, "custom_output.wav")
if err != nil {
    fmt.Printf("语音合成失败: %v\n", err)
    return
}

fmt.Printf("音频数据大小: %d 字节\n", len(result.AudioData))
if result.Subtitle != "" {
    fmt.Printf("字幕信息: %s\n", result.Subtitle)
}
```

## API 参考

### TTSConfig

语音合成客户端配置结构体：

```go
type TTSConfig struct {
    AccessKeyID     string // 阿里云访问密钥ID
    AccessKeySecret string // 阿里云访问密钥Secret
    AppKey          string // 应用AppKey
    URL             string // 服务地址，默认使用阿里云公有云地址
    Region          string // 区域，默认cn-shanghai
}
```

### TTSOptions

语音合成选项结构体：

```go
type TTSOptions struct {
    Voice          string // 发音人，默认xiaoyun
    Format         string // 音频格式，默认WAV
    SampleRate     int    // 采样率，默认16000
    Volume         int    // 音量，范围0-100，默认50
    SpeechRate     int    // 语速，范围-500到500，默认0
    PitchRate      int    // 音调，范围-500到500，默认0
    EnableSubtitle bool   // 是否开启字幕功能
}
```

### 主要方法

#### TextToSpeech

```go
func (c *TTSClient) TextToSpeech(ctx context.Context, text string, options *TTSOptions, outputPath string) (*TTSResult, error)
```

通用的文本转语音方法，支持保存文件和返回音频数据。

#### TextToSpeechToBytes

```go
func (c *TTSClient) TextToSpeechToBytes(ctx context.Context, text string, options *TTSOptions) ([]byte, error)
```

将文本转换为语音并返回音频字节数据。

#### TextToSpeechToFile

```go
func (c *TTSClient) TextToSpeechToFile(ctx context.Context, text string, options *TTSOptions, filePath string) error
```

将文本转换为语音并保存到指定文件。

## 支持的发音人

常用发音人列表：

- `xiaoyun` - 小云（女声，默认）
- `xiaogang` - 小刚（男声）
- `xiaomeng` - 小梦（女声）
- `xiaoxue` - 小雪（女声）
- `xiaofeng` - 小峰（男声）

更多发音人请参考阿里云官方文档。

## 支持的音频格式

- `WAV` - WAV格式（默认）
- `PCM` - PCM格式
- `OPUS` - OPUS格式
- `OPU` - OPU格式

## 错误处理

客户端提供了完善的错误处理机制：

```go
result, err := client.TextToSpeech(ctx, text, options, outputPath)
if err != nil {
    switch {
    case err == context.DeadlineExceeded:
        fmt.Println("请求超时")
    case err == context.Canceled:
        fmt.Println("请求被取消")
    default:
        fmt.Printf("语音合成失败: %v\n", err)
    }
    return
}
```

## 注意事项

1. **认证信息安全**：请妥善保管您的 AccessKey ID 和 AccessKey Secret，不要在代码中硬编码。
2. **网络超时**：建议设置合适的上下文超时时间，避免长时间等待。
3. **文本长度限制**：单次合成的文本长度有限制，具体请参考阿里云官方文档。
4. **并发控制**：客户端是线程安全的，但建议控制并发数量以避免触发限流。
5. **Token缓存**：SDK内部会自动处理Token的获取和缓存，无需手动管理。

## 许可证

本项目基于阿里云官方SDK，请遵循相应的许可证条款。