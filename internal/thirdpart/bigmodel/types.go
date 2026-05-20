package bigmodel

import "time"

// Message 消息结构体
type Message struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"` // 消息内容
}

// ChatCompletionRequest 对话补全请求
type ChatCompletionRequest struct {
	Model       string    `json:"model"`                 // 模型名称，如 "glm-4", "glm-4v", "glm-3-turbo"
	Messages    []Message `json:"messages"`              // 对话消息列表
	Temperature *float64  `json:"temperature,omitempty"` // 温度参数，控制随机性 0-1
	TopP        *float64  `json:"top_p,omitempty"`       // 核采样参数 0-1
	Stream      bool      `json:"stream,omitempty"`      // 是否流式返回
	MaxTokens   *int      `json:"max_tokens,omitempty"`  // 最大生成token数
	Stop        []string  `json:"stop,omitempty"`        // 停止词列表
}

// ChatCompletionResponse 对话补全响应
type ChatCompletionResponse struct {
	ID      string                 `json:"id"`      // 请求ID
	Object  string                 `json:"object"`  // 对象类型
	Created int64                  `json:"created"` // 创建时间戳
	Model   string                 `json:"model"`   // 使用的模型
	Choices []ChatCompletionChoice `json:"choices"` // 生成的选择列表
	Usage   Usage                  `json:"usage"`   // token使用情况
}

// ChatCompletionChoice 对话补全选择项
type ChatCompletionChoice struct {
	Index        int     `json:"index"`          // 选择项索引
	Message      Message `json:"message"`        // 生成的消息
	FinishReason string  `json:"finish_reason"`  // 结束原因
}

// ChatCompletionStreamResponse 流式对话补全响应
type ChatCompletionStreamResponse struct {
	ID      string                       `json:"id"`      // 请求ID
	Object  string                       `json:"object"`  // 对象类型
	Created int64                        `json:"created"` // 创建时间戳
	Model   string                       `json:"model"`   // 使用的模型
	Choices []ChatCompletionStreamChoice `json:"choices"` // 流式选择列表
}

// ChatCompletionStreamChoice 流式对话补全选择项
type ChatCompletionStreamChoice struct {
	Index int                      `json:"index"`         // 选择项索引
	Delta ChatCompletionStreamDelta `json:"delta"`         // 增量内容
	FinishReason *string           `json:"finish_reason"` // 结束原因
}

// ChatCompletionStreamDelta 流式响应增量内容
type ChatCompletionStreamDelta struct {
	Role    string `json:"role,omitempty"`    // 角色
	Content string `json:"content,omitempty"` // 内容增量
}

// ImageGenerationRequest 图像生成请求
type ImageGenerationRequest struct {
	Model  string `json:"model"`            // 模型名称，如 "cogview-3"
	Prompt string `json:"prompt"`           // 图像描述提示词
	Size   string `json:"size,omitempty"`   // 图像尺寸，如 "1024x1024"
	N      *int   `json:"n,omitempty"`      // 生成图像数量，默认1
	Quality string `json:"quality,omitempty"` // 图像质量，如 "standard", "hd"
}

// ImageGenerationResponse 图像生成响应
type ImageGenerationResponse struct {
	Created int64       `json:"created"` // 创建时间戳
	Data    []ImageData `json:"data"`    // 生成的图像数据列表
}

// ImageData 图像数据
type ImageData struct {
	URL           string `json:"url,omitempty"`            // 图像URL
	B64JSON       string `json:"b64_json,omitempty"`       // Base64编码的图像数据
	RevisedPrompt string `json:"revised_prompt,omitempty"` // 修订后的提示词
}

// Usage token使用情况
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`     // 提示词token数
	CompletionTokens int `json:"completion_tokens"` // 生成内容token数
	TotalTokens      int `json:"total_tokens"`      // 总token数
}

// StreamCallback 流式响应回调函数类型
type StreamCallback func(response ChatCompletionStreamResponse) error

// TTSRequest TTS文本转语音请求
type TTSRequest struct {
	Model             string  `json:"model"`                        // 模型名称，固定为 "cogtts"
	Input             string  `json:"input"`                        // 要转换的文本，最大4096字符
	Voice             string  `json:"voice"`                        // 音色选择：tongtong(默认), chuichui, xiaochen, jam, kazi, douji, luodo
	Speed             *float64 `json:"speed,omitempty"`              // 语速，范围[0.5, 2]，默认1.0
	Volume            *float64 `json:"volume,omitempty"`             // 音量，范围(0, 10]，默认1.0
	ResponseFormat    *string  `json:"response_format,omitempty"`    // 音频格式，默认wav
	WatermarkEnabled  *bool    `json:"watermark_enabled,omitempty"`  // 是否添加水印，默认true
}

// TTSVoice 支持的音色类型
type TTSVoice string

const (
	TTSVoiceTongtong TTSVoice = "tongtong" // 默认音色
	TTSVoiceChuichui TTSVoice = "chuichui"
	TTSVoiceXiaochen TTSVoice = "xiaochen"
	TTSVoiceJam      TTSVoice = "jam"
	TTSVoiceKazi     TTSVoice = "kazi"
	TTSVoiceDouji    TTSVoice = "douji"
	TTSVoiceLuodo    TTSVoice = "luodo"
)

// TTSResponseFormat 支持的音频格式
type TTSResponseFormat string

const (
	TTSResponseFormatWAV TTSResponseFormat = "wav" // 默认格式
	TTSResponseFormatMP3 TTSResponseFormat = "mp3"
)

// TTSResponse TTS文本转语音响应（直接返回音频数据）
type TTSResponse struct {
	Data        []byte `json:"-"`          // 音频文件的二进制数据
	ContentType string `json:"-"`          // 内容类型，如 "audio/wav" 或 "audio/mpeg"
	Filename    string `json:"-"`          // 建议的文件名
}

// TTSError TTS错误响应
type TTSError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// RequestOptions 请求选项
type RequestOptions struct {
	Timeout time.Duration // 请求超时时间
}