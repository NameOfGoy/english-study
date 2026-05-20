package wordexample

// PromptTemplate 定义了生成例句的提示模板
type PromptTemplate struct {
	Template   string // 提示模板
	JSONFormat string // JSON格式模板
}

// DefaultPromptTemplate 默认的提示模板
var DefaultPromptTemplate = PromptTemplate{
	Template: "请生成关于词语 {{.Word}} 的例句，包括中文释义。" +
		"{{if .Pos}}其中例句的该单词的词性为 {{.Pos}}，{{end}}" +
		"{{if .Translation}}并指定该单词的中文释义为[{{.Translation}}]，{{end}}" +
		"例句数量为 {{.Count}}。" +
		"如果该单词存在或者指定了多种释义，那么例句就根据给定的例句数量挑取不同的释义生成对应的指定数量的例句。" +
		"返回格式采用json，json的结构为：{{.JSONFormat}}, 不要加任何```或者\n之类的修饰, 就纯json字符串的回复, 否则我的程序json反序列化会失败",
	JSONFormat: "[{\"example\": \"\",\"translation\": \"\"}]",
}

// TemplateData 模板数据
type TemplateData struct {
	Word        string // 单词
	Pos         string // 词性
	Translation string // 指定的单词中文释义
	Count       int    // 例句数量
	JSONFormat  string // JSON格式
}
