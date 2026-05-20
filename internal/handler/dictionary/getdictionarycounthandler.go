package dictionary

import (
	"english-study/internal/errors"
	"english-study/internal/utils"
	"net/http"

	"english-study/internal/logic/dictionary"
	"english-study/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetDictionaryCountHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ui, err := utils.GetUserInfoFromCtx(r.Context())
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(nil, errors.ErrorRequestParamError("获取用户信息失败").WithCause(err)))
			return
		}

		l := dictionary.NewGetDictionaryCountLogic(r.Context(), svcCtx, ui)
		resp, err := l.GetDictionaryCount()
		httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(resp, err))
	}
}
