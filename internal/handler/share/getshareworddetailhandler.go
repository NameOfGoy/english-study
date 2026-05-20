package share

import (
	"english-study/internal/errors"
	"english-study/internal/logic/share"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

func GetShareWordDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetShareWordDetailReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(nil, errors.ErrorRequestParamError("参数解析错误").WithCause(err)))
			return
		}

		// 从上下文获取用户信息
		ui, err := utils.GetUserInfoFromCtx(r.Context())
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(nil, errors.ErrorRequestParamError("获取用户信息失败").WithCause(err)))
			return
		}

		l := share.NewGetShareWordDetailLogic(r.Context(), svcCtx, ui)
		resp, err := l.GetShareWordDetail(&req)
		httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(resp, err))
	}
}
