package config

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// ViperConfig 基于Viper的配置管理器
type ViperConfig struct {
	viper *viper.Viper
}

// AppConfig 应用配置结构体
type AppConfig struct {
	WordPicturePromptTemplate  string `mapstructure:"WordPicturePromptTemplate"`
	WordPhoneticPromptTemplate string `mapstructure:"WordPhoneticPromptTemplate"`
}

// NewViperConfig 创建新的Viper配置管理器
// 支持两种方式：
// 1. 分别传入路径和文件名：NewViperConfig("etc", "englishstudy-api")
// 2. 传入完整路径：NewViperConfig("etc/englishstudy-api.yaml", "")
func NewViperConfig(configPath, configName string) (*ViperConfig, error) {
	v := viper.New()

	// 如果configName为空，则认为configPath是完整的文件路径
	if configName == "" {
		v.SetConfigFile(configPath)
	} else {
		// 设置配置文件路径和名称
		v.AddConfigPath(configPath)
		v.SetConfigName(configName)
		v.SetConfigType("yaml")
	}

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	return &ViperConfig{viper: v}, nil
}

// GetWordPicturePromptTemplate 获取单词图片生成提示模板
func (vc *ViperConfig) GetWordPicturePromptTemplate() string {
	return vc.viper.GetString("WordPicturePromptTemplate")
}

// GetWordPhoneticPromptTemplate 获取单词音标生成提示模板
func (vc *ViperConfig) GetWordPhoneticPromptTemplate() string {
	return vc.viper.GetString("WordPhoneticPromptTemplate")
}

// GetAppConfig 获取完整的应用配置
func (vc *ViperConfig) GetAppConfig() (*AppConfig, error) {
	var config AppConfig
	if err := vc.viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("配置绑定失败: %v", err)
	}
	return &config, nil
}

// WatchConfig 启用配置文件热更新
func (vc *ViperConfig) WatchConfig(callback func()) {
	vc.viper.WatchConfig()
	vc.viper.OnConfigChange(func(e fsnotify.Event) {
		log.Printf("配置文件被修改: %s", e.Name)
		if callback != nil {
			callback()
		}
	})
}

// 全局配置管理器实例
var GlobalViperConfig *ViperConfig

// InitGlobalConfig 初始化全局配置
// 支持两种方式：
// 1. 分别传入路径和文件名：InitGlobalConfig("etc", "englishstudy-api")
// 2. 传入完整路径：InitGlobalConfig("etc/englishstudy-api.yaml", "")
func InitGlobalConfig(configPath, configName string) error {
	var err error
	GlobalViperConfig, err = NewViperConfig(configPath, configName)
	if err != nil {
		return fmt.Errorf("初始化全局配置失败: %v", err)
	}

	// 启用热更新
	GlobalViperConfig.WatchConfig(func() {
		log.Println("配置已重新加载")
	})

	return nil
}

// GetGlobalWordPicturePromtTemplate 获取全局单词图片生成提示模板
func GetGlobalWordPicturePromtTemplate() string {
	if GlobalViperConfig == nil {
		log.Println("警告: 全局配置未初始化，返回默认模板")
		return "请为单词 {{.Word}} 生成一张图片，图片应该能够直观地表达这个单词的含义。"
	}
	return GlobalViperConfig.GetWordPicturePromptTemplate()
}

// GetGlobalWordPhoneticPromptTemplate 获取全局单词音标生成提示模板
func GetGlobalWordPhoneticPromptTemplate() string {
	if GlobalViperConfig == nil {
		log.Println("警告: 全局配置未初始化，返回默认模板")
		return "你是一个专业的英语语音学专家，请为给定的英语单词生成准确的国际音标(IPA)。\n\n要求：\n1. 只返回国际音标，不要包含其他内容\n2. 使用标准的IPA符号\n3. 如果单词有多种发音，请提供最常用的发音\n4. 音标格式：/音标内容/\n\n单词：{{.Word}}\n口音：{{.Accent}}\n\n请生成该单词的国际音标："
	}
	return GlobalViperConfig.GetWordPhoneticPromptTemplate()
}
