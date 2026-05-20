package alibabacloud

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	nls "github.com/aliyun/alibabacloud-nls-go-sdk"
)

// TTSConfig 阿里云语音合成配置
type TTSConfig struct {
	// 阿里云访问密钥ID
	AccessKeyID string
	// 阿里云访问密钥Secret
	AccessKeySecret string
	// 应用AppKey
	AppKey string
	// 服务地址，默认使用阿里云公有云地址
	URL string
	// 区域，默认cn-shanghai
	Region string
}

// TokenInfo Token信息
type TokenInfo struct {
	// Token值
	Token string
	// 过期时间戳（秒）
	ExpireTime int64
	// 获取时间
	ObtainTime time.Time
}

// TTSClient 阿里云语音合成客户端
type TTSClient struct {
	config    *TTSConfig
	logger    *nls.NlsLogger
	tokenInfo *TokenInfo
	tokenMu   sync.RWMutex // Token读写锁
}

// TTSOptions 语音合成选项
type TTSOptions struct {
	// 发音人，默认xiaoyun
	Voice string
	// 音频格式，默认WAV
	Format string
	// 采样率，默认16000
	SampleRate int
	// 音量，范围0-100，默认50
	Volume int
	// 语速，范围-500到500，默认0
	SpeechRate int
	// 音高，范围-500到500，默认0
	PitchRate int
	// 是否开启字幕功能
	EnableSubtitle bool
}

// TTSResult 语音合成结果
type TTSResult struct {
	// 音频数据
	AudioData []byte
	// 字幕信息（如果开启）
	Subtitle string
	// 错误信息
	Error error
}

// NewTTSClient 创建新的语音合成客户端
func NewTTSClient(config *TTSConfig) *TTSClient {
	if config.URL == "" {
		config.URL = nls.DEFAULT_URL
	}
	if config.Region == "" {
		config.Region = "cn-shanghai"
	}
	logger := nls.DefaultNlsLog()
	logger.SetDebug(false) // 不打debug
	logger.SetLogSil(true) // info也不打
	return &TTSClient{
		config: config,
		logger: logger,
	}
}

// getValidToken 获取有效的Token，如果Token过期或不存在则自动刷新
func (c *TTSClient) getValidToken(ctx context.Context) (string, error) {
	c.tokenMu.RLock()
	currentToken := c.tokenInfo
	c.tokenMu.RUnlock()

	// 检查Token是否存在且未过期（提前30秒刷新）
	if currentToken != nil && time.Now().Unix() < (currentToken.ExpireTime-30) {
		return currentToken.Token, nil
	}

	// 需要获取新Token
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	// 双重检查，防止并发获取Token
	if c.tokenInfo != nil && time.Now().Unix() < (c.tokenInfo.ExpireTime-30) {
		return c.tokenInfo.Token, nil
	}

	// 获取新Token
	tokenInfo, err := c.refreshToken(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}

	c.tokenInfo = tokenInfo
	return tokenInfo.Token, nil
}

// refreshToken 刷新Token
func (c *TTSClient) refreshToken(ctx context.Context) (*TokenInfo, error) {
	// 使用GetToken函数获取Token
	tokenResult, err := nls.GetToken(
		nls.DEFAULT_DISTRIBUTE,   // 区域
		nls.DEFAULT_DOMAIN,       // 域名
		c.config.AccessKeyID,     // AKID
		c.config.AccessKeySecret, // AKKEY
		nls.DEFAULT_VERSION,      // 版本
	)
	if err != nil {
		return nil, fmt.Errorf("获取Token失败: %v", err)
	}

	// 更新Token信息
	return &TokenInfo{
		Token:      tokenResult.TokenResult.Id,
		ExpireTime: tokenResult.TokenResult.ExpireTime,
		ObtainTime: time.Now(),
	}, nil
}

// getDefaultOptions 获取默认的语音合成选项
func (c *TTSClient) getDefaultOptions() *TTSOptions {
	return &TTSOptions{
		Voice:          "xiaoyun",
		Format:         "mp3",
		SampleRate:     16000,
		Volume:         50,
		SpeechRate:     0,
		PitchRate:      0,
		EnableSubtitle: false,
	}
}

// mergeOptions 合并选项，用户选项覆盖默认选项
func (c *TTSClient) mergeOptions(userOptions *TTSOptions) *TTSOptions {
	options := c.getDefaultOptions()
	if userOptions == nil {
		return options
	}

	if userOptions.Voice != "" {
		options.Voice = userOptions.Voice
	}
	if userOptions.Format != "" {
		options.Format = userOptions.Format
	}
	if userOptions.SampleRate > 0 {
		options.SampleRate = userOptions.SampleRate
	}
	if userOptions.Volume >= 0 && userOptions.Volume <= 100 {
		options.Volume = userOptions.Volume
	}
	if userOptions.SpeechRate >= -500 && userOptions.SpeechRate <= 500 {
		options.SpeechRate = userOptions.SpeechRate
	}
	if userOptions.PitchRate >= -500 && userOptions.PitchRate <= 500 {
		options.PitchRate = userOptions.PitchRate
	}
	options.EnableSubtitle = userOptions.EnableSubtitle

	return options
}

// TextToSpeech 将文本转换为语音
// text: 要合成的文本内容
// options: 语音合成选项，可以为nil使用默认选项
// outputPath: 输出文件路径，如果为空则只返回音频数据不保存文件
func (c *TTSClient) TextToSpeech(ctx context.Context, text string, options *TTSOptions, outputPath string) (*TTSResult, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	// 合并选项
	finalOptions := c.mergeOptions(options)

	// 获取有效Token
	token, err := c.getValidToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get valid token: %w", err)
	}

	// 使用Token创建连接配置
	connectionConfig := nls.NewConnectionConfigWithToken(c.config.URL, c.config.AppKey, token)

	// 创建结果容器
	result := &TTSResult{
		AudioData: make([]byte, 0),
	}

	// 用于同步的等待组和互斥锁
	var wg sync.WaitGroup
	var mu sync.Mutex
	var synthesisError error

	// 任务失败回调
	taskFailed := func(message string, param interface{}) {
		mu.Lock()
		defer mu.Unlock()
		synthesisError = fmt.Errorf("synthesis failed: %s", message)
		wg.Done()
	}

	// 语音合成数据回调
	synthesisResult := func(data []byte, param interface{}) {
		mu.Lock()
		defer mu.Unlock()
		result.AudioData = append(result.AudioData, data...)
	}

	// 字幕数据回调
	metaInfo := func(subtitle string, param interface{}) {
		mu.Lock()
		defer mu.Unlock()
		result.Subtitle += subtitle
	}

	// 合成完成回调
	completed := func(message string, param interface{}) {
		wg.Done()
	}

	// 连接关闭回调
	closed := func(param interface{}) {
		// 连接关闭处理
	}

	// 创建语音合成对象
	tts, err := nls.NewSpeechSynthesis(
		connectionConfig,
		c.logger,
		false, // 短文本模式
		taskFailed,
		synthesisResult,
		metaInfo,
		completed,
		closed,
		nil, // 用户自定义参数
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create speech synthesis: %w", err)
	}

	// 设置语音合成参数
	param := nls.SpeechSynthesisStartParam{
		Voice:          finalOptions.Voice,
		Format:         finalOptions.Format,
		SampleRate:     finalOptions.SampleRate,
		Volume:         finalOptions.Volume,
		SpeechRate:     finalOptions.SpeechRate,
		PitchRate:      finalOptions.PitchRate,
		EnableSubtitle: finalOptions.EnableSubtitle,
	}

	// 开始合成
	wg.Add(1)
	_, err = tts.Start(text, param, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start synthesis: %w", err)
	}

	// 等待合成完成或超时
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		tts.Shutdown()
		return nil, ctx.Err()
	case <-done:
		// 合成完成
		tts.Shutdown()
	case <-time.After(30 * time.Second):
		tts.Shutdown()
		return nil, fmt.Errorf("synthesis timeout")
	}

	// 检查是否有错误
	if synthesisError != nil {
		return nil, synthesisError
	}

	// 如果指定了输出路径，保存文件
	if outputPath != "" {
		err = c.saveAudioFile(result.AudioData, outputPath)
		if err != nil {
			return result, fmt.Errorf("failed to save audio file: %w", err)
		}
	}

	return result, nil
}

// saveAudioFile 保存音频文件
func (c *TTSClient) saveAudioFile(audioData []byte, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(audioData)
	if err != nil {
		return fmt.Errorf("failed to write audio data: %w", err)
	}

	return nil
}

// TextToSpeechToFile 将文本转换为语音并保存到文件
func (c *TTSClient) TextToSpeechToFile(ctx context.Context, text string, options *TTSOptions, filePath string) error {
	_, err := c.TextToSpeech(ctx, text, options, filePath)
	return err
}

// TextToSpeechToBytes 将文本转换为语音并返回音频字节数据
func (c *TTSClient) TextToSpeechToBytes(ctx context.Context, text string, options *TTSOptions) ([]byte, error) {
	result, err := c.TextToSpeech(ctx, text, options, "")
	if err != nil {
		return nil, err
	}
	return result.AudioData, nil
}
