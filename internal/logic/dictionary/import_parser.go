package dictionary

import (
	"strings"
)

// ImportItem 是导入文件解析后的一行待处理项.
//
//	Word     - 单词或短语原文 (已 trim)
//	TagNames - 这条词要打的标签名列表 (已去重, 顺序: 段标签 → 行末内联标签 → ...)
type ImportItem struct {
	Word     string
	TagNames []string
}

// importMarkerTerminator 是显式的"清标签"标记. 严格字面量匹配, 大小写敏感.
// 设计 §9 确认: 任何其他 [xxx] 都按标签名处理, 哪怕长得像 [---  ] 也不算.
const importMarkerTerminator = "---"

// ParseImportLines 把导入文件每行 (已分行, 未必 trim) 解析成 ImportItem 序列.
//
// 行格式:
//   - 纯空行: 跳过
//   - 整行就是 [X] 一个标记 (无其它文字): 段标记
//       X == "---" → 清当前段标签; 后续词条若无内联标签则无标签
//       X != "---" → 设当前段标签为 X (隐式结束上一个段)
//   - 其它非空行: 词条 + 可选的"行末内联标签"
//       从右往左 peel [tag] 标记, 剩下的就是词条 (Word).
//       例: "look up [重点] [考试]" → Word="look up", inline=[重点, 考试]
//       例: "apple[t1][t2]" → Word="apple", inline=[t1, t2]
//       该词条最终的 TagNames = (段标签, 如有) ∪ (内联标签, 按出现顺序), 去重.
//
// 嵌套段标签不允许 (一次只有一个 "当前段"); 但内联标签可以多个.
func ParseImportLines(lines []string) []ImportItem {
	out := make([]ImportItem, 0, len(lines))
	currentSection := ""

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		word, inlineTags := peelInlineTags(line)
		if word == "" {
			// 整行只有一个 [X] 标记 (一定也只有一个, 因为 peel 后 word 才会空)
			// → 段标记: terminator 清空, 否则设为新段
			if len(inlineTags) == 1 {
				if inlineTags[0] == importMarkerTerminator {
					currentSection = ""
				} else {
					currentSection = inlineTags[0]
				}
				continue
			}
			// 多个 [X][Y] 没词条 → 当作连续段切换, 最后一个生效 (与多个独立标记行等价)
			last := inlineTags[len(inlineTags)-1]
			if last == importMarkerTerminator {
				currentSection = ""
			} else {
				currentSection = last
			}
			continue
		}
		// 普通词条
		tagNames := combineTagNames(currentSection, inlineTags)
		out = append(out, ImportItem{Word: word, TagNames: tagNames})
	}
	return out
}

// peelInlineTags 从行的右端开始反复剥离 [xxx] 标记, 直到剥不动为止.
// 返回剩余的词条文本 (已 trim) + 标签名列表 (按 LEFT-TO-RIGHT 出现顺序).
//
//	"apple [t1] [t2]" → ("apple", ["t1", "t2"])
//	"look up[t1][t2]" → ("look up", ["t1", "t2"])
//	"[only-tag]"      → ("", ["only-tag"])      // 给段标记分支用
//	"apple"           → ("apple", nil)
//	"apple [bad"      → ("apple [bad", nil)     // 不闭合, 不剥
func peelInlineTags(line string) (string, []string) {
	var tags []string
	for {
		rest, name, ok := peelOneTrailingTag(line)
		if !ok {
			break
		}
		// 反向收集; 后面再翻转保证左到右顺序
		tags = append(tags, name)
		line = strings.TrimRight(rest, " \t")
	}
	// 翻转
	for i, j := 0, len(tags)-1; i < j; i, j = i+1, j-1 {
		tags[i], tags[j] = tags[j], tags[i]
	}
	return line, tags
}

// peelOneTrailingTag 如果 line 以一个完整的 [xxx] 结尾, 返回 (前缀, "xxx", true).
// 否则返回 (line, "", false).
func peelOneTrailingTag(line string) (string, string, bool) {
	if len(line) < 2 || line[len(line)-1] != ']' {
		return line, "", false
	}
	// 找最右边匹配的 [, 要求 [ 后内容非空且不含 ] (避免嵌套 [a[b]] 误判)
	for i := len(line) - 2; i >= 0; i-- {
		c := line[i]
		if c == ']' {
			return line, "", false
		}
		if c == '[' {
			inner := line[i+1 : len(line)-1]
			if inner == "" {
				return line, "", false
			}
			return line[:i], inner, true
		}
	}
	return line, "", false
}

// combineTagNames 段标签放最前, 然后内联标签按出现顺序; 整体去重 (大小写敏感).
func combineTagNames(section string, inline []string) []string {
	if section == "" && len(inline) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(inline)+1)
	out := make([]string, 0, len(inline)+1)
	add := func(n string) {
		if n == "" {
			return
		}
		if _, ok := seen[n]; ok {
			return
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	add(section)
	for _, n := range inline {
		add(n)
	}
	return out
}

// parseMarker 已被 peelInlineTags 替代, 保留是为了不破坏其它包对该符号的可能引用 (虽然目前没有).
// 判断 line 是否整体形如 [xxx]; 不再用于 ParseImportLines.
func parseMarker(line string) (string, bool) {
	if len(line) < 3 {
		return "", false
	}
	if line[0] != '[' || line[len(line)-1] != ']' {
		return "", false
	}
	inner := line[1 : len(line)-1]
	if inner == "" {
		return "", false
	}
	// 内部不能含 ] (避免 [a]b] 这种被误认作 marker)
	if strings.Contains(inner, "]") {
		return "", false
	}
	return inner, true
}
