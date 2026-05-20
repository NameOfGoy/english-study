package bigmodel

import (
	"context"
	"time"
)

// SDK 智谱AI SDK主结构体
type SDK struct {
	client       *Client
	chatService  *ChatService
	imageService *ImageService
	ttsService   *TTSService
	auth         *BearerTokenAuth
}

// New 创建新的智谱AI SDK实例
func New(apiKey string) *SDK {
	client := NewClient(apiKey)
	auth := NewBearerTokenAuth(apiKey)

	return &SDK{
		client:       client,
		chatService:  NewChatService(client),
		imageService: NewImageService(client),
		ttsService:   NewTTSService(client),
		auth:         auth,
	}
}

// NewWithConfig 使用自定义配置创建SDK实例
func NewWithConfig(apiKey, baseURL string, timeout time.Duration) *SDK {
	client := NewClientWithConfig(apiKey, baseURL, timeout)
	auth := NewBearerTokenAuth(apiKey)

	return &SDK{
		client:       client,
		chatService:  NewChatService(client),
		imageService: NewImageService(client),
		ttsService:   NewTTSService(client),
		auth:         auth,
	}
}

// Chat 获取对话服务
func (sdk *SDK) Chat() *ChatService {
	return sdk.chatService
}

// Image 获取图像服务
func (sdk *SDK) Image() *ImageService {
	return sdk.imageService
}

// TTS 获取TTS服务
func (sdk *SDK) TTS() *TTSService {
	return sdk.ttsService
}

// SetTimeout 设置请求超时时间
func (sdk *SDK) SetTimeout(timeout time.Duration) {
	sdk.client.SetTimeout(timeout)
	sdk.auth.SetTimeout(timeout)
}

// SetAPIKey 设置API密钥
func (sdk *SDK) SetAPIKey(apiKey string) {
	sdk.client.apiKey = apiKey
	sdk.auth.SetAPIKey(apiKey)
}

// ValidateAPIKey 验证API密钥
func (sdk *SDK) ValidateAPIKey() error {
	return sdk.auth.ValidateAPIKey()
}

// GetAPIKey 获取API密钥（遮蔽显示）
func (sdk *SDK) GetAPIKey() string {
	return MaskAPIKey(sdk.auth.GetAPIKey())
}

// 便捷方法 - 对话相关

// SimpleChat 简单对话
func (sdk *SDK) SimpleChat(ctx context.Context, model, message string) (string, error) {
	resp, err := sdk.chatService.CreateSimpleChat(ctx, model, message)
	if err != nil {
		return "", err
	}
	return GetChatResponse(resp), nil
}

// ChatWithSystem 带系统提示的对话
func (sdk *SDK) ChatWithSystem(ctx context.Context, model, systemMessage, userMessage string) (string, error) {
	resp, err := sdk.chatService.CreateChatWithSystem(ctx, model, systemMessage, userMessage)
	if err != nil {
		return "", err
	}
	return GetChatResponse(resp), nil
}

// StreamChat 流式对话
func (sdk *SDK) StreamChat(ctx context.Context, model, message string, callback StreamCallback) error {
	req := ChatCompletionRequest{
		Model: model,
		Messages: []Message{
			{
				Role:    "user",
				Content: message,
			},
		},
	}
	return sdk.chatService.CreateChatCompletionStream(ctx, req, callback)
}

// 便捷方法 - 图像相关

// GenerateImage 生成图像
func (sdk *SDK) GenerateImage(ctx context.Context, model, prompt string) (string, error) {
	resp, err := sdk.imageService.CreateSimpleImage(ctx, model, prompt)
	if err != nil {
		return "", err
	}
	return GetFirstImageURL(resp), nil
}

// GenerateImageWithSize 生成指定尺寸的图像
func (sdk *SDK) GenerateImageWithSize(ctx context.Context, model, prompt, size string) (string, error) {
	resp, err := sdk.imageService.CreateImageWithSize(ctx, model, prompt, size)
	if err != nil {
		return "", err
	}
	return GetFirstImageURL(resp), nil
}

// GenerateMultipleImages 生成多张图像
func (sdk *SDK) GenerateMultipleImages(ctx context.Context, model, prompt string, count int) ([]string, error) {
	resp, err := sdk.imageService.CreateMultipleImages(ctx, model, prompt, count)
	if err != nil {
		return nil, err
	}
	return GetImageURLs(resp), nil
}

// GenerateHDImage 生成高清图像
func (sdk *SDK) GenerateHDImage(ctx context.Context, model, prompt string) (string, error) {
	resp, err := sdk.imageService.CreateHDImage(ctx, model, prompt)
	if err != nil {
		return "", err
	}
	return GetFirstImageURL(resp), nil
}

// 常用模型常量
const (
	// 对话模型
	ModelGLM45              = "glm-4.5"              // GLM-4.5 高智能旗舰模型 (收费) - 性能最优，强大的推理能力、代码生成能力以及工具调用能力
	ModelGLM45X             = "glm-4.5-x"            // GLM-4.5-X 高智能旗舰极速版模型 (收费) - 推理速度更快，适用于搜索问答、智能助手、实时翻译等时效性较强场景
	ModelGLM45Air           = "glm-4.5-air"          // GLM-4.5-Air 高性价比模型 (收费) - 同参数规模性能最佳，在推理、编码和智能体任务上表现强劲
	ModelGLM45AirX          = "glm-4.5-airx"         // GLM-4.5-AirX 高性价比极速版模型 (收费) - 推理速度快且价格适中，适用于时效性有较强要求的场景
	ModelGLM4Plus           = "glm-4-plus"           // GLM-4-Plus 性能优秀模型 (收费) - 性能最优，语言理解、逻辑推理、指令遵循、长文本处理效果领先
	ModelGLM4Air250414      = "glm-4-air-250414"    // GLM-4-Air-250414 高性价比模型 (收费) - 快速执行复杂任务，擅长工具调用、联网搜索、代码
	ModelGLM4Long           = "glm-4-long"           // GLM-4-Long 超长输入模型 (收费) - 支持高达1M的上下文长度，能够理解和回应复杂的查询，专为处理超长文本和记忆型任务设计
	ModelGLM4AirX           = "glm-4-airx"           // GLM-4-AirX 极速推理模型 (收费) - 超快的推理速度，强大的推理效果
	ModelGLM4FlashX250414   = "glm-4-flashx-250414" // GLM-4-FlashX-250414 高速低价模型 (收费) - Flash增强版本，超快推理速度，更快并发保障
	ModelGLMZ1Air           = "glm-z1-air"           // GLM-Z1-Air 高性价比模型 (收费) - 高性价比，具备深度思考能力，数理推理能力显著增强
	ModelGLMZ1AirX          = "glm-z1-airx"          // GLM-Z1-AirX 极速推理模型 (收费) - 国内最快的推理速度，支持8倍推理速度，问题即问即答
	ModelGLMZ1FlashX        = "glm-z1-flashx"        // GLM-Z1-FlashX 高速低价模型 (收费) - 超快推理速度，更快并发保障，极致性价比
	ModelGLM45Flash         = "glm-4.5-flash"        // GLM-4.5-Flash 模型 (免费) - 最新基座模型的普惠版本
	ModelGLM4Flash250414    = "glm-4-flash-250414"  // GLM-4-Flash-250414 模型 (免费) - 超长上下文处理能力，多语言支持，支持外部工具调用
	ModelGLMZ1Flash         = "glm-z1-flash"         // GLM-Z1-Flash 模型 (免费) - 复杂任务推理，轻量级应用

	// 视觉模型
	ModelGLM45V     = "glm-4.5v"     // GLM-4.5V 旗舰视觉推理模型 (收费) - 同级别开源视觉推理模型SOTA，覆盖视频理解、复杂文件解析、前端复刻等核心场景，新增思考模式开关
	ModelGLM41VThinkingFlashX = "glm-4.1v-thinking-flashx" // GLM-4.1V-Thinking-FlashX 轻量视觉推理模型 (收费) - 视觉推理能力，复杂场景理解，多步骤分析，高并发
	ModelGLM4VPlus0111 = "glm-4v-plus-0111" // GLM-4V-Plus-0111 视觉理解模型 (收费) - 图像理解能力强，支持图文对话，场景识别准确
	ModelGLM41VThinkingFlash = "glm-4.1v-thinking-flash" // GLM-4.1V-Thinking-Flash 视觉推理模型 (免费) - 视觉推理能力，复杂场景理解，多步骤分析
	ModelGLM4VFlash = "glm-4v-flash" // GLM-4V-Flash 视觉理解模型 (免费) - 图像理解，多语言支持

	// 图像生成模型
	ModelCogView3   = "cogview-3"    // CogView-3 图像生成模型 (收费)
	ModelCogView4   = "cogview-4"    // CogView-4 图像生成模型 (收费) - 高质量图像生成，风格多样化，细节丰富
	ModelCogView3Flash = "cogview-3-flash" // CogView-3-Flash 图像生成模型 (免费) - 创意丰富多样，推理速度快

	// 视频生成模型
	ModelCogVideoX3 = "cogvideox-3" // CogVideoX-3 高智能旗舰视频生成模型 (收费) - 主观清晰度大幅提升，更好的指令遵循、物理真实模拟，现实、3D风格场景表现提升，新增首尾帧生成功能
	ModelCogVideoX2 = "cogvideox-2" // CogVideoX-2 高性价比视频生成模型 (收费) - 支持主体进行大幅度运动，驾驭多种艺术风格
	ModelViduQ1 = "vidu-q1" // Vidu Q1 质量较优视频生成模型 (收费) - 影视级的画质清晰度，精准解决画面崩坏，多艺术形态的风格，行业标杆级转场流畅度
	ModelVidu2 = "vidu-2" // Vidu 2 高速低价视频生成模型 (收费) - 速度优、性价比优，语义增强的首尾帧衔接，多参考图的一致性强化
	ModelCogVideoXFlash = "cogvideox-flash" // CogVideoX-Flash 视频生成模型 (免费) - 沉浸式AI音效，4K高清画质呈现，10秒视频时长拓展，60fps高帧率输出

	// 音视频模型
	ModelGLM4Voice = "glm-4-voice" // GLM-4-Voice 语音模型 (收费) - 直接理解和生成中英文语音，实现实时语音对话，根据用户指令灵活调整语音的情感、语调、语速和方言等特性
	ModelGLMRealtime = "glm-realtime" // GLM-Realtime 实时音视频模型 (收费) - 能够提供实时的视频通话功能，通话记忆时长长达2分钟，具有跨文本、音频和视频进行实时推理的能力
	ModelGLMASR = "glm-asr" // GLM-ASR 语音识别模型 (收费) - 上下文智能理解，强抗噪性能，多语言多方言覆盖

	// 常用图像尺寸
	ImageSize256    = "256x256"
	ImageSize512    = "512x512"
	ImageSize1024   = "1024x1024"
	ImageSize1024x1792 = "1024x1792"
	ImageSize1792x1024 = "1792x1024"

	// 图像质量
	ImageQualityStandard = "standard"
	ImageQualityHD       = "hd"
)

// 版本信息
const (
	SDKVersion = "1.0.0"
	APIVersion = "v4"
)