package oss

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// SafeKeyPart 把任意用户输入安全化为 OSS 对象 key 的一段路径片段.
// 规则:
//   - 全小写后, 只保留 [a-z0-9_-]; 其它字符替换成 '-'
//   - 结果若与原串不同 (说明含非安全字符), 追加 sha256 前 8 字符避免冲突
//   - 空串返回 "_"
//
// 为什么不直接 hash 整体: 保留可读性方便排查 OSS 目录.
func SafeKeyPart(s string) string {
	if s == "" {
		return "_"
	}
	lower := strings.ToLower(s)
	var b strings.Builder
	b.Grow(len(lower))
	for _, r := range lower {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '_', r == '-':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	cleaned := b.String()
	if cleaned == lower {
		return cleaned
	}
	sum := sha256.Sum256([]byte(s))
	return cleaned + "_" + hex.EncodeToString(sum[:4])
}
