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

func BindWxHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserBindWxReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(nil, errors.ErrorRequestParamError("参数解析错误").WithCause(err)))
			return
		}

		// BindWx 是公开接口(用户此时尚未登录), 不从 JWT 上下文取 ui;
		// 身份由请求体里的 account/password 校验。注意: 默认模板会给带请求体的 handler 加 ui 提取,
		// 公开接口需手动去掉(goctl 跳过已存在 handler, 本改动可保留)。
		l := user.NewBindWxLogic(r.Context(), svcCtx)
		resp, err := l.BindWx(&req)
		httpx.OkJsonCtx(r.Context(), w, utils.WrapResponse(resp, err))
	}
}
