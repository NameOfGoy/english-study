package types

import (
	"encoding/json"
	"strings"
)

func (w *WordPhrase) ExampleString() string {
	if len(w.Example) == 0 {
		return "" // 避免 nil → "null" 字符串
	}
	data, _ := json.Marshal(w.Example)
	return string(data)
}

func (w *WordPhrase) ExampleObject(s string) error {
	s = strings.TrimSpace(s)
	if s == "" || s == "null" {
		w.Example = nil
		return nil
	}
	return json.Unmarshal([]byte(s), &w.Example)
}
