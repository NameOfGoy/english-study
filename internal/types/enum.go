package types

const (
	OssBucket = "englishstudy"
)

const (
	WordPosUnknown          = iota
	WordPosNoun             // 名词
	WordPosTransitiveVerb   // 及物动词
	WordPosIntransitiveVerb // 不及物动词
	WordPosAdverb           // 副词
	WordPosAdjectives       // 形容词
	WordPosPreposition      // 介词
	WordPosConjunction      // 连词
	WordPosInterjection     // 感叹词
	WordPosParticle         // 助词
	WordPosPronoun          // 代词
	WordPosNumber           // 数词
	WordPosArticle          // 冠词
	WordPosAuxiliaryVerb    // 辅助动词
)

// WordPosMap 词性的中文名称
var WordPosChineseMap = map[int]string{
	WordPosUnknown:          "未知",
	WordPosNoun:             "名词",
	WordPosTransitiveVerb:   "及物动词",
	WordPosIntransitiveVerb: "不及物动词",
	WordPosAdverb:           "副词",
	WordPosAdjectives:       "形容词",
	WordPosPreposition:      "介词",
	WordPosConjunction:      "连词",
	WordPosInterjection:     "感叹词",
	WordPosParticle:         "助词",
	WordPosPronoun:          "代词",
	WordPosNumber:           "数词",
	WordPosArticle:          "冠词",
	WordPosAuxiliaryVerb:    "辅助动词",
}

func ToPosChinese(i int) string {
	if v, ok := WordPosChineseMap[i]; ok {
		return v
	}
	return WordPosChineseMap[WordPosUnknown]
}

// WordPosSwMap 词性的英文缩写
var WordPosSwMap = map[int]string{
	WordPosUnknown:          "unknown",
	WordPosNoun:             "n.",
	WordPosTransitiveVerb:   "vt.",
	WordPosIntransitiveVerb: "vi.",
	WordPosAdverb:           "adv.",
	WordPosAdjectives:       "adj.",
	WordPosPreposition:      "prep.",
	WordPosConjunction:      "conj.",
	WordPosInterjection:     "interj.",
	WordPosParticle:         "part.",
	WordPosPronoun:          "pron.",
	WordPosNumber:           "num.",
	WordPosArticle:          "art.",
	WordPosAuxiliaryVerb:    "aux.",
}

func ToPosSw(i int) string {
	if v, ok := WordPosSwMap[i]; ok {
		return v
	}
	return WordPosSwMap[WordPosUnknown]
}

// 缩写的枚举映射
var WordPosSwToEnumMap = map[string]int{
	"unknown": WordPosUnknown,
	"n.":      WordPosNoun,
	"vt.":     WordPosTransitiveVerb,
	"vi.":     WordPosIntransitiveVerb,
	"adv.":    WordPosAdverb,
	"adj.":    WordPosAdjectives,
	"prep.":   WordPosPreposition,
	"conj.":   WordPosConjunction,
	"interj.": WordPosInterjection,
	"part.":   WordPosParticle,
	"pron.":   WordPosPronoun,
	"num.":    WordPosNumber,
	"art.":    WordPosArticle,
	"aux.":    WordPosAuxiliaryVerb,
}

func WordPosSwToEnum(sw string) int {
	switch sw {
	case "a.":
		sw = "adj."
	case "v.":
		sw = "vt."
	}

	if v, ok := WordPosSwToEnumMap[sw]; ok {
		return v
	}
	return WordPosUnknown
}

/*
Exchange 变化形式
p	过去式（did）
d	过去分词（done）
i	现在分词（doing）
3	第三人称单数（does）
r	形容词比较级（-er）
t	形容词最高级（-est）
s	名词复数形式
*/
const (
	WordPosExchangePast     = "p"
	WordPosExchangePastPart = "d"
	WordPosExchangePresent  = "i"
	WordPosExchange3rdSing  = "3"
	WordPosExchangeAdjComp  = "r"
	WordPosExchangeAdjSuper = "t"
	WordPosExchangeNounPlur = "s"
)

// 变化形式的中文名称
var WordPosExchangeChineseMap = map[string]string{
	WordPosExchangePast:     "过去式",
	WordPosExchangePastPart: "过去分词",
	WordPosExchangePresent:  "现在分词",
	WordPosExchange3rdSing:  "第三人称单数",
	WordPosExchangeAdjComp:  "形容词比较级",
	WordPosExchangeAdjSuper: "形容词最高级",
	WordPosExchangeNounPlur: "名词复数形式",
}

// ToExchangeChinese 获取变化形式的中文名称
func ToExchangeChinese(i string) string {
	if v, ok := WordPosExchangeChineseMap[i]; ok {
		return v
	}
	return ""
}

// 词性所对应的变化形式
var WordPosExchangeMap = map[int][]string{
	WordPosTransitiveVerb:   {WordPosExchangePast, WordPosExchangePastPart, WordPosExchangePresent, WordPosExchange3rdSing},
	WordPosIntransitiveVerb: {WordPosExchangePast, WordPosExchangePastPart, WordPosExchangePresent, WordPosExchange3rdSing},
	WordPosAdjectives:       {WordPosExchangeAdjComp, WordPosExchangeAdjSuper},
	WordPosNoun:             {WordPosExchangeNounPlur},
}

// PosExchange 获取词性的变化形式
func PosExchange(pos int) []string {
	if v, ok := WordPosExchangeMap[pos]; ok {
		return v
	}
	return nil
}

// 词语类型
const (
	WordTypeUnknown = iota // 未知类型
	WordTypeWord           // 单词
	WordTypePhrase         // 短语
)

/*
单词状态定义
*/
const (
	WordStatusUnknown    = iota // 未知标签
	WordStatusStudy             // 学习状态
	WordStatusReview            // 复习状态
	WordStatusStrengthen        // 强化状态
	WordStatusFinish            // 完成状态
)
