package file

import (
	"io"
	"net/http"
	"strings"
)

// ProxyImageHandler proxies an external image URL to avoid CORS issues in the browser.
// Usage: GET /api/v1/file-service/proxy-image?url=https://example.com/image.jpg
func ProxyImageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		imageURL := r.URL.Query().Get("url")
		if imageURL == "" {
			http.Error(w, "missing url parameter", http.StatusBadRequest)
			return
		}

		// Basic validation: only allow http/https URLs
		if !strings.HasPrefix(imageURL, "http://") && !strings.HasPrefix(imageURL, "https://") {
			http.Error(w, "invalid url", http.StatusBadRequest)
			return
		}

		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, imageURL, nil)
		if err != nil {
			http.Error(w, "failed to create request", http.StatusInternalServerError)
			return
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "failed to fetch image", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Forward content type
		contentType := resp.Header.Get("Content-Type")
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		} else {
			w.Header().Set("Content-Type", "image/jpeg")
		}
		w.Header().Set("Cache-Control", "public, max-age=86400")

		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}
