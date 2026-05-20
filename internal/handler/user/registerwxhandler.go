package user

import (
	"english-study/internal/errors"
	"english-study/internal/logic/user"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

func RegisterWxHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserRegisterWxReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(nil, errors.ErrorRequestParamError("参数解析错误").WithCause(err)))
			return
		}

		l := user.NewRegisterWxLogic(r.Context(), svcCtx)
		resp, err := l.RegisterWx(&req)
		httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(resp, err))
	}
}
