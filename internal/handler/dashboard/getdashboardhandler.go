package dashboard

import (
	"net/http"

	"english-study/internal/errors"
	"english-study/internal/logic/dashboard"
	"english-study/internal/svc"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetDashboardHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ui, err := utils.GetUserInfoFromCtx(r.Context())
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(nil, errors.ErrorRequestParamError("获取用户信息失败").WithCause(err)))
			return
		}

		l := dashboard.NewGetDashboardLogic(r.Context(), svcCtx, ui)
		resp, err := l.GetDashboard()
		httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(resp, err))
	}
}
