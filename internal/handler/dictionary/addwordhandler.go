package dictionary

import (
	"english-study/internal/errors"
	"english-study/internal/logic/dictionary"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func AddWordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AddWordReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(nil, errors.ErrorRequestParamError("参数解析错误").WithCause(err)))
			return
		}

		ui, err := utils.GetUserInfoFromCtx(r.Context())
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(nil, errors.ErrorRequestParamError("获取用户信息错误").WithCause(err)))
			return
		}

		l := dictionary.NewAddWordLogic(r.Context(), svcCtx, ui)
		resp, err := l.AddWord(&req)
		httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(resp, err))
	}
}
