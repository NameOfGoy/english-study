package cc

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"time"

	"english-study/internal/ccauth"
	"english-study/internal/errors"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type RelayLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewRelayLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *RelayLoginLogic {
	return &RelayLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

// RelayLogin 二次门禁登录:
//   1) 必须 admin role (前置 JWT 已校过, 这里防御性再校)
//   2) IP 错密钥 5 次 / 5min 锁定
//   3) AccessKey constant-time 比对
//   4) 通过则签发 access + refresh, 把 refresh jti 加 store
func (l *RelayLoginLogic) RelayLogin(req *types.RelayLoginReq) (*types.RelayLoginResp, error) {
	if err := utils.RequireAdmin(l.ui); err != nil {
		return nil, errors.ErrorPermissionError("仅管理员可登录 AI 辅助")
	}

	ip := ccauth.ClientIPFromCtx(l.ctx)
	if ip != "" {
		if allowed, remaining := l.svcCtx.CCIPLimiter.Allowed(ip); !allowed {
			logx.WithContext(l.ctx).Errorf("ai 辅助登录被锁: ip=%s 还需 %ds", ip, remaining)
			return nil, errors.ErrorPermissionError("尝试次数过多, 请稍后再试")
		}
	}

	expectedKey := l.svcCtx.Config.CC.AccessKey
	if expectedKey == "" {
		return nil, errors.ErrorTokenGenerateError("CC.AccessKey 未配置, 联系运维")
	}
	if subtle.ConstantTimeCompare([]byte(req.AccessKey), []byte(expectedKey)) != 1 {
		if ip != "" {
			l.svcCtx.CCIPLimiter.RecordFailure(ip)
		}
		// sleep 1s 拖慢爆破, 跟前端 UX 之间的平衡; 真要爆破的人也不会因 sleep 1s 收手, 主要靠 5 次锁
		time.Sleep(time.Second)
		return nil, errors.ErrorPermissionError("AI 辅助密钥错误")
	}
	if ip != "" {
		l.svcCtx.CCIPLimiter.RecordSuccess(ip)
	}

	return l.issueTokens()
}

// issueTokens 跟 refresh logic 共用: 签发新 access + 新 refresh + 把新 refresh jti 加 store.
// 返回供给 handler.
func (l *RelayLoginLogic) issueTokens() (*types.RelayLoginResp, error) {
	cfg := l.svcCtx.Config
	accessTTL := cfg.CC.AccessTokenTTL
	if accessTTL <= 0 {
		accessTTL = 900 // 15min
	}
	refreshTTL := cfg.CC.RefreshTokenTTL
	if refreshTTL <= 0 {
		refreshTTL = 7 * 24 * 3600 // 7 天
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

// newJTI 16 字节随机 hex, 32 字符
func newJTI() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
