package middleware

import "net/http"

// SecurityHeaders 给所有响应加几个常用的安全头.
// 不上 CSP — 当前前端依赖 inline style 较多, CSP 收紧需要单独评估.
func SecurityHeaders(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		// 浏览器不要 MIME-sniff 上传内容
		h.Set("X-Content-Type-Options", "nosniff")
		// 不允许被任何 iframe 嵌套, 防 clickjacking
		h.Set("X-Frame-Options", "DENY")
		// 跨站只发送 origin, 不带路径 (减少 referer 内含 token 等的泄漏)
		h.Set("Referrer-Policy", "no-referrer")
		// 旧 XSS Auditor (现代浏览器已废弃, 但开了也不亏)
		h.Set("X-XSS-Protection", "0")
		next(w, r)
	}
}
