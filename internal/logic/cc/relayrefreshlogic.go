package cc

import (
	"context"
	"fmt"
	"time"

	"english-study/internal/errors"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zeromicro/go-zero/core/logx"
)

type RelayRefreshLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewRelayRefreshLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *RelayRefreshLogic {
	return &RelayRefreshLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

// RelayRefresh 用 refresh token 换新 access + 新 refresh (rotation):
//   1) admin role (前置 JWT 已校过)
//   2) 解析+验签 refresh token: aud=forwarder, token_type=refresh, jti 在 store
//   3) revoke 旧 jti (rotation 关键; 防止 refresh 被盗后两端都能用)
//   4) 签发新 access + 新 refresh, 新 jti 入 store
//
// 如果 refresh 用户身份跟 JWT context 里 admin 身份不符, 拒绝 (防止 A 的 refresh 让 B 用)
func (l *RelayRefreshLogic) RelayRefresh(req *types.RelayRefreshReq) (*types.RelayLoginResp, error) {
	if err := utils.RequireAdmin(l.ui); err != nil {
		return nil, errors.ErrorPermissionError("仅管理员可刷新 AI 辅助 token")
	}
	if req.RefreshToken == "" {
		return nil, errors.ErrorRequestParamError("refresh_token 不能为空")
	}

	claims, err := l.parseRefresh(req.RefreshToken)
	if err != nil {
		return nil, errors.ErrorTokenParseError("refresh token 无效").WithCause(err)
	}

	// 身份一致性: refresh token 里的 user_id 必须等于当前 JWT context 里的 admin id
	refreshUID, _ := claims["user_id"].(float64)
	if uint(refreshUID) != l.ui.ID {
		l.svcCtx.CCRefreshStore.Revoke(claimsStr(claims, "jti")) // 顺手作废这条可疑 refresh
		return nil, errors.ErrorPermissionError("refresh token 跟当前登录身份不符")
	}

	oldJTI := claimsStr(claims, "jti")
	if oldJTI == "" || !l.svcCtx.CCRefreshStore.IsValid(oldJTI) {
		return nil, errors.ErrorTokenParseError("refresh token 已失效, 请重新输 AI 辅助密钥")
	}
	// rotation: 旧 jti 立刻撤销, 防止重放
	l.svcCtx.CCRefreshStore.Revoke(oldJTI)

	// 复用 issueTokens (logic 实际是 RelayLoginLogic 上的方法; 这里 inline 一份避免 cross-struct 调用)
	return l.issue()
}

func (l *RelayRefreshLogic) issue() (*types.RelayLoginResp, error) {
	cfg := l.svcCtx.Config
	accessTTL := cfg.CC.AccessTokenTTL
	if accessTTL <= 0 {
		accessTTL = 900
	}
	refreshTTL := cfg.CC.RefreshTokenTTL
	if refreshTTL <= 0 {
		refreshTTL = 7 * 24 * 3600
	}
	wsURL := cfg.CC.RelayWSURL
	if wsURL == "" {
		return nil, errors.ErrorTokenGenerateError("CC.RelayWSURL 未配置")
	}

	iat := time.Now().Unix()
	jti, err := newJTI()
	if err != nil {
		return nil, errors.ErrorTokenGenerateError("生成 jti 失败").WithCause(err)
	}
	access, err := utils.GenerateRelayAccessToken(cfg.Auth.AccessSecret, iat, accessTTL, l.ui.ID, l.ui.Username, l.ui.Role, jti)
	if err != nil {
		return nil, errors.ErrorTokenGenerateError("签发 access token 失败").WithCause(err)
	}
	refresh, err := utils.GenerateRelayRefreshToken(cfg.Auth.AccessSecret, iat, refreshTTL, l.ui.ID, l.ui.Username, l.ui.Role, jti)
	if err != nil {
		return nil, errors.ErrorTokenGenerateError("签发 refresh token 失败").WithCause(err)
	}
	l.svcCtx.CCRefreshStore.Add(jti, iat+refreshTTL)

	return &types.RelayLoginResp{
		AccessToken:      access,
		RefreshToken:     refresh,
		AccessExpiresAt:  iat + accessTTL,
		RefreshExpiresAt: iat + refreshTTL,
		WsURL:            wsURL,
	}, nil
}

func (l *RelayRefreshLogic) parseRefresh(token string) (jwt.MapClaims, error) {
	var claims jwt.MapClaims
	t, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected alg: %v", t.Method.Alg())
		}
		return []byte(l.svcCtx.Config.Auth.AccessSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}
	if !t.Valid {
		return nil, fmt.Errorf("invalid")
	}
	// aud
	auds, _ := claims["aud"]
	switch v := auds.(type) {
	case string:
		if v != utils.RelayAudience {
			return nil, fmt.Errorf("aud mismatch")
		}
	case []any:
		ok := false
		for _, a := range v {
			if s, _ := a.(string); s == utils.RelayAudience {
				ok = true
				break
			}
		}
		if !ok {
			return nil, fmt.Errorf("aud mismatch")
		}
	default:
		return nil, fmt.Errorf("aud type %T", auds)
	}
	// token_type 必须是 refresh
	tt, _ := claims["token_type"].(string)
	if tt != utils.RelayTokenTypeRefresh {
		return nil, fmt.Errorf("token_type=%q, expect refresh (access token 不能调 /relay-refresh)", tt)
	}
	return claims, nil
}

func claimsStr(c jwt.MapClaims, k string) string {
	if v, ok := c[k].(string); ok {
		return v
	}
	return ""
}
