package cc

import (
	"net/http"

	"english-study/internal/errors"
	"english-study/internal/logic/cc"
	"english-study/internal/svc"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// UploadFileHandler AI 桥文件上传 (手动注册路由, 不走 goctl; multipart 不适合 goctl typed handler).
// 路由在 englishstudy.go 用 server.AddRoutes + WithJwt 注册, 所以这里 ctx 已带 JWT claims.
func UploadFileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(ccMaxMultipartMem); err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(nil, errors.ErrorRequestParamError("参数解析错误").WithCause(err)))
			return
		}
		ui, err := utils.GetUserInfoFromCtx(r.Context())
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(nil, errors.ErrorPermissionError("未登录").WithCause(err)))
			return
		}
		f, fh, err := r.FormFile("file")
		if f != nil {
			defer f.Close()
		}
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(nil, errors.ErrorRequestParamError("缺少文件").WithCause(err)))
			return
		}

		l := cc.NewUploadFileLogic(r.Context(), svcCtx, ui)
		resp, err := l.UploadFile(f, fh)
		httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(resp, err))
	}
}

// ccMaxMultipartMem multipart 解析时驻留内存上限 (超出落临时文件); 10MB 够, 大文件由 logic 的 20MB 上限把关.
const ccMaxMultipartMem = 10 << 20
