package utils

import "strings"

func GetOptionFromOptions[OT any, OF ~func(op *OT)](ops []OF, init ...func() OT) OT {
	var op OT
	if len(init) > 0 {
		op = init[0]()
	}
	for _, o := range ops {
		o(&op)
	}
	return op
}

func TrimMarkDownJsonTag(content string) string {
	// 去掉```json```标签
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimSuffix(content, "```")
	// 去掉首尾空格
	content = strings.TrimSpace(content)
	return content
}
