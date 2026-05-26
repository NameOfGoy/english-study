package utils

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-do-not-use"

func TestGenerateToken_EmbedsRole(t *testing.T) {
	tests := []struct {
		name string
		role int
	}{
		{"normal user", RoleNormal},
		{"admin user", RoleAdmin},
		{"unknown role kept verbatim", 99},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tok, err := GenerateToken(testSecret, time.Now().Unix(), 3600, 42, "alice", tt.role)
			if err != nil {
				t.Fatalf("GenerateToken: %v", err)
			}
			// 解出来验证 claim
			parsed, err := jwt.Parse(tok, func(t *jwt.Token) (any, error) {
				return []byte(testSecret), nil
			})
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			claims, ok := parsed.Claims.(jwt.MapClaims)
			if !ok {
				t.Fatalf("claims type %T", parsed.Claims)
			}
			gotRole, ok := claims["role"].(float64) // JSON number → float64 by default
			if !ok {
				t.Fatalf("role claim missing or not float64: %T", claims["role"])
			}
			if int(gotRole) != tt.role {
				t.Errorf("role got %d want %d", int(gotRole), tt.role)
			}
		})
	}
}

func TestGenerateToken_TamperFailsVerify(t *testing.T) {
	tok, err := GenerateToken(testSecret, time.Now().Unix(), 3600, 42, "alice", RoleNormal)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	// 用错误的 secret 解 → 签名错
	_, err = jwt.Parse(tok, func(t *jwt.Token) (any, error) { return []byte("wrong"), nil })
	if err == nil {
		t.Fatal("expected signature error, got nil")
	}
}

func TestUserInfo_IsAdmin(t *testing.T) {
	tests := []struct {
		name string
		ui   *UserInfo
		want bool
	}{
		{"nil receiver safe", nil, false},
		{"normal", &UserInfo{Role: RoleNormal}, false},
		{"admin", &UserInfo{Role: RoleAdmin}, true},
		{"weird role not admin", &UserInfo{Role: 99}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ui.IsAdmin(); got != tt.want {
				t.Errorf("IsAdmin got %v want %v", got, tt.want)
			}
		})
	}
}

func TestRequireAdmin(t *testing.T) {
	if err := RequireAdmin(&UserInfo{Role: RoleAdmin}); err != nil {
		t.Errorf("admin should pass, got %v", err)
	}
	if err := RequireAdmin(&UserInfo{Role: RoleNormal}); err == nil {
		t.Error("normal should fail, got nil")
	}
	if err := RequireAdmin(nil); err == nil {
		t.Error("nil should fail, got nil")
	}
}

// ctxWithClaim 模拟 go-zero JWT 中间件把 claims 塞进 ctx 的方式 (json.Number 数字类型)
func ctxWithClaim(uid int, username string, role *int) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "user_id", json.Number(itoa(uid)))
	ctx = context.WithValue(ctx, "username", username)
	if role != nil {
		ctx = context.WithValue(ctx, "role", json.Number(itoa(*role)))
	}
	return ctx
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func TestGetUserInfoFromCtx_ReadsRole(t *testing.T) {
	r := RoleAdmin
	ctx := ctxWithClaim(7, "bob", &r)
	ui, err := GetUserInfoFromCtx(ctx)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if ui.ID != 7 || ui.Username != "bob" || ui.Role != RoleAdmin {
		t.Errorf("got %+v want id=7 username=bob role=admin", ui)
	}
}

func TestGetUserInfoFromCtx_MissingRoleDefaultsToNormal(t *testing.T) {
	// 老 JWT 没 role claim, 必须 fallback 0 不能 panic
	ctx := ctxWithClaim(7, "bob", nil)
	ui, err := GetUserInfoFromCtx(ctx)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if ui.Role != RoleNormal {
		t.Errorf("expected role=normal for missing claim, got %d", ui.Role)
	}
}

func TestGetUserInfoFromCtx_MissingUserIDFails(t *testing.T) {
	ctx := context.Background()
	if _, err := GetUserInfoFromCtx(ctx); err == nil {
		t.Error("expected err for missing user_id, got nil")
	}
}
