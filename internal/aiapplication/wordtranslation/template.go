package wordtranslation

// PromptTemplate 定义了生成翻译的提示模板
type PromptTemplate struct {
	Template string // 提示模板
}

// DefaultPromptTemplate 默认的提示模板（纯中文文本输出）
var DefaultPromptTemplate = PromptTemplate{
	Template: "请为英语词语 {{.Word}} 生成一个常见且准确的中文释义。" +
		"{{if .Pos}}该词语的词性为 {{.Pos}}。{{end}}" +
		"如果该短语不是一个合法的英语短语，或者该短语没有中文释义，又或者短语中存在拼写错误，那么请返回代码块里的字符串```错误：$错误原因```" +
		"直接返回纯中文释义文本，不要包含任何 JSON、代码块、引号或额外说明，只返回释义或者错误本身。我用来做词典用的，所以只需要返回中文释义，不需要其他说明。",
}

// TemplateData 模板数据
type TemplateData struct {
	Word string // 单词/词语
	Pos  string // 词性
}
