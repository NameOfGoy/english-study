package alibabacloud

import (
	"context"
	"reflect"
	"testing"
	"time"

	nls "github.com/aliyun/alibabacloud-nls-go-sdk"
)

// TestNewTTSClient 测试创建TTS客户端
func TestNewTTSClient(t *testing.T) {
	config := &TTSConfig{
		AccessKeyID:     "test-key-id",
		AccessKeySecret: "test-key-secret",
		AppKey:          "test-app-key",
	}

	client := NewTTSClient(config)
	if client == nil {
		t.Fatal("Failed to create TTS client")
	}

	if client.config.AccessKeyID != "test-key-id" {
		t.Errorf("Expected AccessKeyID to be 'test-key-id', got '%s'", client.config.AccessKeyID)
	}

	// 测试默认值设置
	if client.config.URL == "" {
		t.Error("URL should be set to default value")
	}

	if client.config.Region != "cn-shanghai" {
		t.Errorf("Expected Region to be 'cn-shanghai', got '%s'", client.config.Region)
	}
}

// TestGetDefaultOptions 测试获取默认选项
func TestGetDefaultOptions(t *testing.T) {
	config := &TTSConfig{
		AccessKeyID:     "test-key-id",
		AccessKeySecret: "test-key-secret",
		AppKey:          "test-app-key",
	}

	client := NewTTSClient(config)
	options := client.getDefaultOptions()

	if options.Voice != "xiaoyun" {
		t.Errorf("Expected default voice to be 'xiaoyun', got '%s'", options.Voice)
	}

	if options.SampleRate != 16000 {
		t.Errorf("Expected default sample rate to be 16000, got %d", options.SampleRate)
	}

	if options.Volume != 50 {
		t.Errorf("Expected default volume to be 50, got %d", options.Volume)
	}

	if options.EnableSubtitle != false {
		t.Errorf("Expected default EnableSubtitle to be false, got %v", options.EnableSubtitle)
	}
}

// TestMergeOptions 测试选项合并
func TestMergeOptions(t *testing.T) {
	config := &TTSConfig{
		AccessKeyID:     "test-key-id",
		AccessKeySecret: "test-key-secret",
		AppKey:          "test-app-key",
	}

	client := NewTTSClient(config)

	// 测试nil选项
	merged := client.mergeOptions(nil)
	if merged.Voice != "xiaoyun" {
		t.Errorf("Expected merged voice to be 'xiaoyun', got '%s'", merged.Voice)
	}

	// 测试部分自定义选项
	customOptions := &TTSOptions{
		Voice:  "xiaogang",
		Volume: 80,
	}

	merged = client.mergeOptions(customOptions)
	if merged.Voice != "xiaogang" {
		t.Errorf("Expected merged voice to be 'xiaogang', got '%s'", merged.Voice)
	}

	if merged.Volume != 80 {
		t.Errorf("Expected merged volume to be 80, got %d", merged.Volume)
	}

	// 默认值应该保持不变
	if merged.SampleRate != 16000 {
		t.Errorf("Expected merged sample rate to be 16000, got %d", merged.SampleRate)
	}

	// 测试边界值验证
	invalidOptions := &TTSOptions{
		Volume:     150,   // 超出范围
		SpeechRate: 1000,  // 超出范围
		PitchRate:  -1000, // 超出范围
	}

	merged = client.mergeOptions(invalidOptions)
	// 无效值应该被忽略，使用默认值
	if merged.Volume != 50 {
		t.Errorf("Expected invalid volume to be ignored, got %d", merged.Volume)
	}

	if merged.SpeechRate != 0 {
		t.Errorf("Expected invalid speech rate to be ignored, got %d", merged.SpeechRate)
	}

	if merged.PitchRate != 0 {
		t.Errorf("Expected invalid pitch rate to be ignored, got %d", merged.PitchRate)
	}
}

// TestTextToSpeechValidation 测试TextToSpeech输入验证
func TestTextToSpeechValidation(t *testing.T) {
	config := &TTSConfig{
		AccessKeyID:     "test-key-id",
		AccessKeySecret: "test-key-secret",
		AppKey:          "test-app-key",
	}

	client := NewTTSClient(config)
	ctx := context.Background()

	// 测试空文本
	_, err := client.TextToSpeech(ctx, "", nil, "")
	if err == nil {
		t.Error("Expected error for empty text, got nil")
	}

	if err.Error() != "text cannot be empty" {
		t.Errorf("Expected 'text cannot be empty' error, got '%s'", err.Error())
	}
}

// BenchmarkMergeOptions 性能测试：选项合并
func BenchmarkMergeOptions(b *testing.B) {
	config := &TTSConfig{
		AccessKeyID:     "test-key-id",
		AccessKeySecret: "test-key-secret",
		AppKey:          "test-app-key",
	}

	client := NewTTSClient(config)
	customOptions := &TTSOptions{
		Voice:      "xiaogang",
		Volume:     80,
		SpeechRate: 100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.mergeOptions(customOptions)
	}
}

// ExampleTTSClient_TextToSpeechToBytes 示例：获取音频字节数据
func ExampleTTSClient_TextToSpeechToBytes() {
	config := &TTSConfig{
		AccessKeyID:     "your-access-key-id",
		AccessKeySecret: "your-access-key-secret",
		AppKey:          "your-app-key",
	}

	client := NewTTSClient(config)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	audioData, err := client.TextToSpeechToBytes(ctx, "Hello, World!", nil)
	if err != nil {
		// 处理错误
		return
	}

	// 使用音频数据
	_ = audioData
}

// ExampleTTSClient_TextToSpeechToFile 示例：保存音频到文件
func ExampleTTSClient_TextToSpeechToFile() {
	config := &TTSConfig{
		AccessKeyID:     "your-access-key-id",
		AccessKeySecret: "your-access-key-secret",
		AppKey:          "your-app-key",
	}

	client := NewTTSClient(config)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	options := &TTSOptions{
		Voice:      "xiaogang",
		Volume:     80,
		SpeechRate: 100,
	}

	err := client.TextToSpeechToFile(ctx, "Hello, World!", options, "output.wav")
	if err != nil {
		// 处理错误
		return
	}

	// 文件已保存
}

func TestExampleUsage(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestExampleUsage",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ExampleUsage()
		})
	}
}

func TestNewTTSClient1(t *testing.T) {
	type args struct {
		config *TTSConfig
	}
	tests := []struct {
		name string
		args args
		want *TTSClient
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTTSClient(tt.args.config); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTTSClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSimpleTextToSpeech(t *testing.T) {
	type args struct {
		accessKeyID     string
		accessKeySecret string
		appKey          string
		text            string
		outputFile      string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SimpleTextToSpeech(tt.args.accessKeyID, tt.args.accessKeySecret, tt.args.appKey, tt.args.text, tt.args.outputFile); (err != nil) != tt.wantErr {
				t.Errorf("SimpleTextToSpeech() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTTSClient_TextToSpeech(t *testing.T) {
	type fields struct {
		config    *TTSConfig
		logger    *nls.NlsLogger
		tokenInfo *TokenInfo
	}
	type args struct {
		ctx        context.Context
		text       string
		options    *TTSOptions
		outputPath string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *TTSResult
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &TTSClient{
				config:    tt.fields.config,
				logger:    tt.fields.logger,
				tokenInfo: tt.fields.tokenInfo,
			}
			got, err := c.TextToSpeech(tt.args.ctx, tt.args.text, tt.args.options, tt.args.outputPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("TextToSpeech() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TextToSpeech() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTTSClient_TextToSpeechToBytes(t *testing.T) {
	type fields struct {
		config    *TTSConfig
		logger    *nls.NlsLogger
		tokenInfo *TokenInfo
	}
	type args struct {
		ctx     context.Context
		text    string
		options *TTSOptions
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &TTSClient{
				config:    tt.fields.config,
				logger:    tt.fields.logger,
				tokenInfo: tt.fields.tokenInfo,
			}
			got, err := c.TextToSpeechToBytes(tt.args.ctx, tt.args.text, tt.args.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("TextToSpeechToBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TextToSpeechToBytes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTTSClient_TextToSpeechToFile(t *testing.T) {
	type fields struct {
		config    *TTSConfig
		logger    *nls.NlsLogger
		tokenInfo *TokenInfo
	}
	type args struct {
		ctx      context.Context
		text     string
		options  *TTSOptions
		filePath string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &TTSClient{
				config:    tt.fields.config,
				logger:    tt.fields.logger,
				tokenInfo: tt.fields.tokenInfo,
			}
			if err := c.TextToSpeechToFile(tt.args.ctx, tt.args.text, tt.args.options, tt.args.filePath); (err != nil) != tt.wantErr {
				t.Errorf("TextToSpeechToFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTTSClient_getDefaultOptions(t *testing.T) {
	type fields struct {
		config    *TTSConfig
		logger    *nls.NlsLogger
		tokenInfo *TokenInfo
	}
	tests := []struct {
		name   string
		fields fields
		want   *TTSOptions
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &TTSClient{
				config:    tt.fields.config,
				logger:    tt.fields.logger,
				tokenInfo: tt.fields.tokenInfo,
			}
			if got := c.getDefaultOptions(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDefaultOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTTSClient_getValidToken(t *testing.T) {
	type fields struct {
		config    *TTSConfig
		logger    *nls.NlsLogger
		tokenInfo *TokenInfo
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &TTSClient{
				config:    tt.fields.config,
				logger:    tt.fields.logger,
				tokenInfo: tt.fields.tokenInfo,
			}
			got, err := c.getValidToken(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("getValidToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getValidToken() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTTSClient_mergeOptions(t *testing.T) {
	type fields struct {
		config    *TTSConfig
		logger    *nls.NlsLogger
		tokenInfo *TokenInfo
	}
	type args struct {
		userOptions *TTSOptions
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *TTSOptions
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &TTSClient{
				config:    tt.fields.config,
				logger:    tt.fields.logger,
				tokenInfo: tt.fields.tokenInfo,
			}
			if got := c.mergeOptions(tt.args.userOptions); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTTSClient_refreshToken(t *testing.T) {
	type fields struct {
		config    *TTSConfig
		logger    *nls.NlsLogger
		tokenInfo *TokenInfo
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *TokenInfo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &TTSClient{
				config:    tt.fields.config,
				logger:    tt.fields.logger,
				tokenInfo: tt.fields.tokenInfo,
			}
			got, err := c.refreshToken(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("refreshToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("refreshToken() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTTSClient_saveAudioFile(t *testing.T) {
	type fields struct {
		config    *TTSConfig
		logger    *nls.NlsLogger
		tokenInfo *TokenInfo
	}
	type args struct {
		audioData []byte
		filePath  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &TTSClient{
				config:    tt.fields.config,
				logger:    tt.fields.logger,
				tokenInfo: tt.fields.tokenInfo,
			}
			if err := c.saveAudioFile(tt.args.audioData, tt.args.filePath); (err != nil) != tt.wantErr {
				t.Errorf("saveAudioFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
