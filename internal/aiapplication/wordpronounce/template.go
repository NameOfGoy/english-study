package wordpronounce

// PromptTemplate 音标生成提示模板
type PromptTemplate struct {
	Template string
}

// DefaultPromptTemplate 默认音标生成模板
var DefaultPromptTemplate = PromptTemplate{
	Template: `你是一个专业的英语语音学专家，请为给定的英语单词生成准确的国际音标(IPA)。

要求：
1. 只返回国际音标，不要包含其他内容
2. 使用标准的IPA符号
4. 音标格式：/音标内容/

单词：{{.Word}}
口音：{{.Accent}}

请生成该单词的国际音标`,
}

// TemplateData 模板数据
type TemplateData struct {
	Word   string // 单词
	Accent string // 口音 | uk/us
}