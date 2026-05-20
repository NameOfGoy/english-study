package file

import (
	"english-study/internal/errors"
	"english-study/internal/logic/file"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

func UploadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FileUploadReq
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(nil, errors.ErrorRequestParamError("参数解析错误").WithCause(err)))
			return
		}
		req.Bucket = r.FormValue("bucket")
		req.Object = r.FormValue("object")

		f, fh, err := r.FormFile("file")
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(nil, errors.ErrorRequestParamError("参数解析错误").WithCause(err)))
			return
		}
		defer f.Close()

		l := file.NewUploadLogic(r.Context(), svcCtx)
		resp, err := l.Upload(&req, f, fh)
		httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(resp, err))
	}
}
