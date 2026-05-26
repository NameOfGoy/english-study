package file

import (
	"english-study/internal/errors"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"
	"fmt"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func DownloadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FileDownloadReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(nil, errors.ErrorRequestParamError("参数解析错误").WithCause(err)))
			return
		}
		
		// 直接 301 转发到文件服务（http.Redirect 会按顺序正确写 Location header + status code）
		http.Redirect(w, r, fmt.Sprintf("/api/v1/file/%s", req.Path), http.StatusMovedPermanently)
	}
}
