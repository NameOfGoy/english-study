package file

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// 单张图片代理上限 20MB, 防止上游返回 GB 级响应撑爆内存
const proxyImageMaxBytes = 20 << 20

// 允许代理的图片源域名（包括子域名匹配）。新增图片源需在此列表显式加入。
var proxyImageAllowedHosts = map[string]bool{
	"www.bing.com":         true,
	"tse1.mm.bing.net":     true,
	"tse2.mm.bing.net":     true,
	"tse3.mm.bing.net":     true,
	"tse4.mm.bing.net":     true,
	"th.bing.com":          true,
	"cn.bing.com":          true,
	"www.google.com":       true,
	"upload.wikimedia.org": true,
}

// 自定义 DialContext: 把白名单 host 在 dial 那一刻再验证一次 IP, 闭合 DNS rebinding 窗口
// (先 LookupIPAddr 再 Do 的传统模式会再做一次 DNS, 攻击者可在两次解析之间把记录指向 169.254.169.254)
var proxyImageDialer = &net.Dialer{Timeout: 5 * time.Second}

var proxyImageTransport = &http.Transport{
	DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, err
		}
		for _, ip := range ips {
			if ip.IP.IsLoopback() || ip.IP.IsPrivate() || ip.IP.IsLinkLocalUnicast() || ip.IP.IsLinkLocalMulticast() || ip.IP.IsUnspecified() {
				return nil, errors.New("resolved IP is private / loopback / link-local")
			}
		}
		// 用第一个公网 IP 直连, 不再走系统 DNS 二次解析
		if len(ips) == 0 {
			return nil, errors.New("no IPs resolved")
		}
		return proxyImageDialer.DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
	},
}

var proxyImageClient = &http.Client{
	Timeout:   15 * time.Second,
	Transport: proxyImageTransport,
}

// ProxyImageHandler proxies an external image URL to avoid CORS issues in the browser.
// Usage: GET /api/v1/file-service/proxy-image?url=https://example.com/image.jpg
//
// 安全约束：
//   - 仅允许 http/https
//   - 域名必须在 proxyImageAllowedHosts 白名单内（防 SSRF 打内网 / 云元数据 169.254.169.254）
//   - 解析后的 IP 必须不是私网 / loopback / link-local
//   - 客户端有 15s 超时
func ProxyImageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		imageURL := r.URL.Query().Get("url")
		if imageURL == "" {
			http.Error(w, "missing url parameter", http.StatusBadRequest)
			return
		}

		parsed, err := url.Parse(imageURL)
		if err != nil {
			http.Error(w, "invalid url", http.StatusBadRequest)
			return
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			http.Error(w, "url scheme not allowed", http.StatusBadRequest)
			return
		}
		host := strings.ToLower(parsed.Hostname())
		if !proxyImageAllowedHosts[host] {
			http.Error(w, "host not allowed", http.StatusForbidden)
			return
		}
		// 注: IP 防护现在主要由 transport DialContext 兜底 (DNS rebinding-proof);
		// 这里保留一次 pre-flight 检查给更友好的错误码 (Forbidden 而非 BadGateway)
		if ips, lookupErr := net.DefaultResolver.LookupIPAddr(r.Context(), host); lookupErr == nil {
			for _, ip := range ips {
				if ip.IP.IsLoopback() || ip.IP.IsPrivate() || ip.IP.IsLinkLocalUnicast() || ip.IP.IsLinkLocalMulticast() || ip.IP.IsUnspecified() {
					http.Error(w, "host resolves to private address", http.StatusForbidden)
					return
				}
			}
		}

		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, imageURL, nil)
		if err != nil {
			http.Error(w, "failed to create request", http.StatusInternalServerError)
			return
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

		resp, err := proxyImageClient.Do(req)
		if err != nil {
			http.Error(w, "failed to fetch image", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// 只回传图片 content-type，避免把上游的 text/html 错误页等当图片回给前端
		contentType := resp.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image/") {
			http.Error(w, "upstream did not return an image", http.StatusBadGateway)
			return
		}
		// 提前拦截声称超大的响应, 节省带宽
		if resp.ContentLength > proxyImageMaxBytes {
			http.Error(w, "upstream image too large", http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "public, max-age=86400")

		w.WriteHeader(resp.StatusCode)
		// io.Copy 限长, 防止上游隐瞒 Content-Length 后流式回传 GB 级数据
		_, _ = io.Copy(w, io.LimitReader(resp.Body, proxyImageMaxBytes))
	}
}
