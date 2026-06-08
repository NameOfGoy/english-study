package cc

import (
	"net/http"
	"strings"

	"english-study/internal/ccauth"
	"english-study/internal/errors"
	"english-study/internal/logic/cc"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func RelayLoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RelayLoginReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(nil, errors.ErrorRequestParamError("参数解析错误").WithCause(err)))
			return
		}

		ui, err := utils.GetUserInfoFromCtx(r.Context())
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(nil, errors.ErrorRequestParamError("获取用户信息失败").WithCause(err)))
			return
		}

		ctx := ccauth.WithClientIP(r.Context(), clientIP(r))
		l := cc.NewRelayLoginLogic(ctx, svcCtx, ui)
		resp, err := l.RelayLogin(&req)
		httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(resp, err))
	}
}

// clientIP 优先取 X-Forwarded-For (nginx 反代后真实 IP), 退到 RemoteAddr.
// XFF 取第一个 (最靠左的客户端), 后面的是中间代理.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// RemoteAddr 形如 "1.2.3.4:5678", 砍端口
	addr := r.RemoteAddr
	if i := strings.LastIndexByte(addr, ':'); i > 0 {
		return addr[:i]
	}
	return addr
}
