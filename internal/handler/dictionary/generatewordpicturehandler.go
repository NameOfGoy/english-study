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

func GenerateWordPictureHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GenerateWordPictureReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(nil, errors.ErrorRequestParamError("参数解析错误").WithCause(err)))
			return
		}

		ui, err := utils.GetUserInfoFromCtx(r.Context())
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(nil, errors.ErrorRequestParamError("获取用户信息错误").WithCause(err)))
			return
		}

		l := dictionary.NewGenerateWordPictureLogic(r.Context(), svcCtx, ui)
		resp, err := l.GenerateWordPicture(&req)
		httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(resp, err))
	}
}
