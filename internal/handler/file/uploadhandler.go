package file

import (
	"english-study/internal/errors"
	"english-study/internal/logic/file"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func UploadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FileUploadReq
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(nil, errors.ErrorRequestParamError("参数解析错误").WithCause(err)))
			return
		}

		// 从 JWT 取 user_id —— 上传必须有登录态
		ui, err := utils.GetUserInfoFromCtx(r.Context())
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(nil, errors.ErrorPermissionError("未登录").WithCause(err)))
			return
		}

		bucket := r.FormValue("bucket")
		// bucket 必须在白名单内（防止往任意 bucket 写）
		if bucket != "" && bucket != types.OssBucket {
			httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(nil, errors.ErrorRequestParamError("非法 bucket")))
			return
		}
		if bucket == "" {
			bucket = types.OssBucket
		}
		req.Bucket = bucket

		rawObject := r.FormValue("object")
		// 标准化 + 反路径穿越（path.Clean 会把 "a/../b" 变成 "b"，".."/绝对路径直接拒）
		cleaned := path.Clean("/" + strings.TrimPrefix(rawObject, "/"))
		if cleaned == "." || cleaned == "/" || strings.Contains(cleaned, "..") {
			httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(nil, errors.ErrorRequestParamError("非法 object 路径")))
			return
		}
		cleaned = strings.TrimPrefix(cleaned, "/")
		// 强制 per-user 命名空间，禁止覆盖别人的对象
		req.Object = fmt.Sprintf("upload/%d/%s", ui.ID, cleaned)

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
