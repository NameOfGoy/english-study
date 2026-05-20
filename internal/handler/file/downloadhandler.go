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
		
		//l := file.NewDownloadLogic(r.Context(), svcCtx, ui)
		//resp, err := l.Download(&req)
		w.WriteHeader(http.StatusMovedPermanently) // 直接301转发
		w.Header().Set("Location", fmt.Sprintf("/api/v1/file/%s", req.Path))
		httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(&types.FileDownloadResp{}, nil))
	}
}
