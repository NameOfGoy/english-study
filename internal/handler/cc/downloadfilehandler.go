package cc

import (
	"io"
	"mime"
	"net/http"
	"path/filepath"

	"english-study/internal/svc"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

// DownloadFileHandler AI 桥文件下载 (手动注册, **无 JWT 中间件**: 凭 ?t= 签名 token 自鉴权).
// CC(本地) curl GET /api/v1/cc/download?t=<token> → 校验 token → 从私有桶 cc-uploads 流式吐出.
// 等价应用层 presigned: 私有桶 + token 自鉴权 + token 带 TTL.
func DownloadFileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("t")
		if token == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}
		bucket, object, err := utils.VerifyFileToken(svcCtx.Config.Auth.AccessSecret, token)
		if err != nil {
			http.Error(w, "invalid or expired token", http.StatusForbidden)
			return
		}

		reader, err := svcCtx.Oss.Download(r.Context(), bucket, object)
		if err != nil {
			http.Error(w, "object not found", http.StatusNotFound)
			return
		}
		defer reader.Close()

		ctype := mime.TypeByExtension(filepath.Ext(object))
		if ctype == "" {
			ctype = "application/octet-stream"
		}
		w.Header().Set("Content-Type", ctype)
		w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(object)+"\"")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")

		if _, err := io.Copy(w, reader); err != nil {
			// body 已开始写, 无法改 status code, 只记日志
			logx.WithContext(r.Context()).Errorf("cc download stream copy failed obj=%s: %v", object, err)
		}
	}
}
