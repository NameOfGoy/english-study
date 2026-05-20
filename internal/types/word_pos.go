package types

import (
	"encoding/json"
	"fmt"
	"strings"
)

type WordPoses []*WordPos

func (w WordPoses) Len() int {
	return len(w)
}

func (w WordPoses) Less(i, j int) bool {
	return w[i].Pos < w[j].Pos
}

func (w WordPoses) Swap(i, j int) {
	w[i], w[j] = w[j], w[i]
}

// Chinese 词性的中文名称
func (w *WordPos) Chinese() string {
	if _, ok := WordPosChineseMap[w.Pos]; !ok {
		return WordPosChineseMap[WordPosUnknown]
	}
	return WordPosChineseMap[w.Pos]
}

// Sw 词性的英文缩写
func (w *WordPos) Sw() string {
	if _, ok := WordPosSwMap[w.Pos]; !ok {
		return WordPosSwMap[WordPosUnknown]
	}
	return WordPosSwMap[w.Pos]
}

func (w *WordPos) ExampleString() string {
	if len(w.Example) == 0 {
		return "" // 避免 json.Marshal(nil) 输出 "null" 字符串
	}
	data, _ := json.Marshal(w.Example)
	return string(data)
}

func (w *WordPos) ExampleObject(s string) error {
	s = strings.TrimSpace(s)
	if s == "" || s == "null" {
		w.Example = nil
		return nil
	}
	return json.Unmarshal([]byte(s), &w.Example)
}

func (w *WordPos) ExchangeString() string {
	if w.Exchange == nil {
		return ""
	}
	// 按:分割
	exchange := make([]string, 0, len(w.Exchange))
	for k, v := range w.Exchange {
		exchange = append(exchange, fmt.Sprintf("%s:%s", k, v))
	}
	return fmt.Sprintf("[%s]", strings.Join(exchange, ","))
}

func (w *WordPos) ExchangeObject(s string) {
	if len(s) < 2 {
		return // 没有变化形式
	}
	// 掐头去尾
	s = s[1 : len(s)-1]
	// 按逗号分割
	exchange := strings.Split(s, ",")
	w.Exchange = make(map[string]string)
	for _, e := range exchange {
		if len(e) == 0 {
			continue
		}
		// 按:分割
		exchange := strings.Split(e, ":")
		w.Exchange[exchange[0]] = exchange[1]
	}
	return
}
