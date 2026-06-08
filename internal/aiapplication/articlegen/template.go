package articlegen

// PromptTemplate 文章生成提示模板.
type PromptTemplate struct {
	Template   string // text/template 文本
	JSONFormat string // 期望的 JSON 结构骨架
}

// ItemData 单个词条在 prompt 中的渲染数据.
type ItemData struct {
	Index   int
	Word    string
	TypeStr string // "单词"/"短语"
	Meaning string
	Forms   string
}

// TemplateData prompt 模板数据.
type TemplateData struct {
	Items      []ItemData
	Count      int
	MinSent    int
	MaxSent    int
	Style      string
	Missing    string // 纠错重问时填: 上次缺失/不匹配的原因
	JSONFormat string
}

// DefaultPromptTemplate 默认中文提示.
// 设计取向(按用户反馈): 第一目标是"行文自然流畅", 不再硬性要求"故事". 体裁自适应——
// 适合讲故事就讲故事(更好), 抽象/专业词就写一段连贯的话题短文即可; 底线是读着顺, 不生硬.
// 正文用 raw string(便于多行 + 模板动作), 仅提及代码围栏的那行用双引号串拼接(raw string 不能含反引号).
var DefaultPromptTemplate = PromptTemplate{
	Template: `你是一位英语老师, 也擅长写自然地道的英文短文. 下面给你一组英语词条, 请用它们写一小段"读起来自然, 流畅, 连贯"的英文短文, 帮助英语学习者在语境中记住这些词.

最重要的要求: 行文自然流畅, 整段像一段完整连贯的文字, 而不是"一句一个词"的孤立例句堆砌.

体裁不限, 你自己挑最能让这些词自然融入的形式:
- 这些词适合讲故事或生活小场景, 就写一个有趣好记的小故事(最好);
- 是抽象或专业的词(如医学, 学术词), 就围绕同一个话题写一段连贯通顺的小短文/小科普即可, 不必硬编故事;
- 也可以是一段日常叙述, 一段说明, 或一小段对话 —— 哪种读着最顺就用哪种.
能有趣, 有场景更好; 做不到也没关系, 但一定要连贯, 通顺, 不生硬.

动笔前先想清楚: 围绕什么主题把这些词串起来, 句子之间怎么自然衔接(用 then / so / but / because / and / after that 等连接词, 用代词承接), 让整段是一个整体.
{{if .Missing}}
上一次这些词没被自然用上, 或它的形态在正文里找不到: {{.Missing}} 请重写整篇, 自然地把它们用进去(个别实在融不进的生僻词可少量舍弃, 但要优先用上).
{{end}}
【要求, 按重要性排序】
0. 标题(必填): 给这段文字起一个贴合内容的英文标题填进 title_en, 它的中文填进 title_zh; 两个都必须写, 绝不能留空.
1. 自然连贯(最重要): 整段围绕同一主题, 句子前后有衔接, 读起来是"一段完整的话", 不是互不相关的例句.
2. 尽量把每个给定词条都自然用上(可用其时态/单复数/派生等变化形式, 词义须符合给定释义); 实在难融入的生僻词, 宁可少用也不要硬塞得很别扭.
3. 一句话可以含 0 个, 1 个或多个目标词; 允许有不含目标词的句子让行文更顺.
4. 篇幅 {{.MinSent}} 到 {{.MaxSent}} 句; 句子简短, 除目标词外尽量用常见词.
5. 中英对照: 每个英文句配一句自然通顺, 符合中文习惯的整句意译(不要逐词直译).
6. surfaces: used_words 里每个词给出它在正文里实际出现的字面形式, 必须是某个英文句子里一字不差的子串(含大小写与词形变化). 例如 run 出现 running 和 ran 就写 ["running","ran"].
7. 纯文本: 标题与所有句子都是纯文本, 严禁任何 markdown 或符号标记(如 **加粗**, *斜体*, # 标题, 反引号包裹的代码).

【严禁(读着生硬的根源)】
- 严禁"一个目标词造一句, 句子各自独立, 互不相关"的例句堆砌 —— 这是最不能接受的.
- 严禁把句子写成可以任意打乱顺序也不影响阅读的样子(说明句间没有联系).
- 严禁为了凑词把不相干的内容硬拼到一起, 读着别扭.
- 严禁任何 markdown / 星号 / 反引号 / # 标记.

【这次要用的词】(顺序你自己按行文安排):
{{range .Items}}{{.Word}}({{.TypeStr}}, 释义: {{.Meaning}}{{if .Forms}}, 可接受变形: {{.Forms}}{{end}}); {{end}}

【风格参考】(仅示意"自然流畅 + 中英对照 + surfaces 的写法"; 它的体裁, 词和内容都和你这次不同, 严禁照抄):
{"title_en":"The Lantern on the River","title_zh":"河上的灯笼","sentences":[{"en":"Late at night, Tom found an old paper boat tied to the dock behind his house.","zh":"深夜里, 汤姆在自家屋后的码头发现了一只系着的旧纸船."},{"en":"He set a small lantern inside it and let the little boat drift slowly down the river.","zh":"他在船里放了一盏小灯笼, 让小船顺着河水缓缓漂流而下."},{"en":"But the wind grew stronger, and the faint light almost went out.","zh":"可是风越来越大, 那点微弱的光几乎要熄灭了."},{"en":"Tom was stubborn, so he made up his mind to follow the boat along the bank.","zh":"汤姆很固执, 于是下定决心沿着河岸追着小船跑."},{"en":"At last an old fisherman caught the boat and handed the warm lantern over to him.","zh":"最后, 一位老渔夫捞起了小船, 把那盏暖暖的灯笼递还给了他."}],"used_words":[{"word":"lantern","type":1,"surfaces":["lantern"]},{"word":"drift","type":1,"surfaces":["drift"]},{"word":"faint","type":1,"surfaces":["faint"]},{"word":"stubborn","type":1,"surfaces":["stubborn"]},{"word":"make up one's mind","type":2,"surfaces":["made up his mind"]},{"word":"hand over","type":2,"surfaces":["handed the warm lantern over"]}]}

【输出格式】
` + "只返回如下结构的纯 JSON, 不要加 ```, ```json 等围栏, 不要任何解释或多余文字; title_en 与 title_zh 必须填真实标题不能为空; type 字段单词填 1, 短语填 2:\n" + `{{.JSONFormat}}`,
	JSONFormat: "{\"title_en\":\"\",\"title_zh\":\"\",\"sentences\":[{\"en\":\"\",\"zh\":\"\"}],\"used_words\":[{\"word\":\"\",\"type\":1,\"surfaces\":[\"\"]}]}",
}
