package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// GuestForbiddenCode 游客(只读)尝试写操作时返回的业务码。
// 取一个不与 kratos 小整数错误码(1,2,3...)冲突的值; 前端据此弹"登录后使用"引导, 而非普通错误 toast。
const GuestForbiddenCode = 40300

// GuestGuard 游客只读守卫。
//
// 游客 token(role=RoleGuest) 只允许 GET(只读浏览); 任何写操作(POST/PUT/DELETE)一律拦下,
// 返回 GuestForbiddenCode, 由前端引导去登录/注册。登录/注册族接口放行(游客升级为正式用户的入口)。
//
// 为何自己解 JWT: 本中间件经 server.Use 全局注册, 执行在 go-zero JWT 中间件之前,
// 此刻 ctx 尚无 role claim, 故直接解 Authorization 头取 role。
// 这样也不依赖 goctl 生成的 routes.go, 不怕 make api 覆盖。
func GuestGuard(jwtSecret string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if guestCanPass(r, jwtSecret) {
				next(w, r)
				return
			}
			httpx.OkJsonCtx(r.Context(), w, types.CommonReply{
				Code:   GuestForbiddenCode,
				Msg:    "请登录后使用该功能",
				Reason: "guest_forbidden",
			})
		}
	}
}

// guestCanPass 判断本次请求是否放行(true=放行)。
func guestCanPass(r *http.Request, secret string) bool {
	// 只读方法一律放行
	switch r.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	}
	// 登录/注册/绑定族: 游客借此升级为正式用户, 必须放行
	p := r.URL.Path
	if strings.Contains(p, "/user/login") || strings.Contains(p, "/user/register") {
		return true
	}
	// 取 role: 非游客 / 无 token / 解析失败 → 放行(鉴权交给后续 JWT 中间件 401)
	role, ok := guestRoleFromAuth(r, secret)
	if !ok {
		return true
	}
	return role != utils.RoleGuest
}

// guestRoleFromAuth 从 Authorization: Bearer <jwt> 解出 role claim。
// 仅当 token 合法且解析成功时返回 (role, true); 其余 (0, false)。role 缺失按非游客(0)处理。
func guestRoleFromAuth(r *http.Request, secret string) (int, bool) {
	const prefix = "Bearer "
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, prefix) {
		return 0, false
	}
	tokenStr := strings.TrimSpace(auth[len(prefix):])
	if tokenStr == "" {
		return 0, false
	}
	claims := jwt.MapClaims{}
	tok, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !tok.Valid {
		return 0, false
	}
	switch v := claims["role"].(type) {
	case float64:
		return int(v), true
	case json.Number:
		n, _ := v.Int64()
		return int(n), true
	default:
		return 0, true // role claim 缺失 → 按非游客处理(放行)
	}
}
