package wordpicture

// PromptTemplate 定义了生成图片的提示模板
type PromptTemplate struct {
	Template string // 提示模板
}

// DefaultPromptTemplate 默认的提示模板
var DefaultPromptTemplate = PromptTemplate{
	Template: "请为英语单词'{{.Word}}'{{if .Pos}}(词性:{{.Pos}}){{end}}{{if .Translation}} 释义为'{{.Translation}}'{{end}}生成一张直观、清晰卡通风格的图片，图片应该能够准确表达这个词语的含义。" +
		"注意!!图片中不要出现该词语!!请确保图片内容与词语含义高度相关，图片应该简洁明了，便于英语学习者理解词语含义。",
}

// TemplateData 模板数据
type TemplateData struct {
	Word        string // 词语
	Pos         string // 词性
	Translation string // 释义
}
