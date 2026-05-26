package utils

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
)

// 密码存储有三种历史格式 (新→旧):
//   1. bcrypt: "$2a$12$..." (60字符), 推荐
//   2. PBKDF2: "<hash_hex>:<salt_hex>", 2048 轮 SHA-256, 32 字节 (强度不足)
//   3. 明文: 不含":" 也不以"$2"开头 (远古遗留)
// 登录验证时全部兼容, needUpgrade=true 表示该用户的存储格式过期, 调用方应趁热把新密文写回.
const (
	passwordSaltLen      = 16
	pbkdf2LegacyRounds   = 2048
	pbkdf2LegacyKeySize  = 32
	bcryptCost           = 12 // ~250ms on a 2024 laptop, 用户体验和强度平衡
)

func pbkdf2Encode(password string, salt []byte) string {
	dk := pbkdf2.Key([]byte(password), salt, pbkdf2LegacyRounds, pbkdf2LegacyKeySize, sha256.New)
	return hex.EncodeToString(dk)
}

// PasswordEncode legacy export, 保留给老调用方; 新代码不要再用.
func PasswordEncode(password string, salt []byte) string {
	return pbkdf2Encode(password, salt)
}

// HashPassword 用 bcrypt(cost=12) 生成密文.
func HashPassword(password string) (string, error) {
	if len(password) > 72 { // bcrypt 输入上限
		return "", fmt.Errorf("密码长度超过 72 字节")
	}
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash 失败: %w", err)
	}
	return string(h), nil
}

// VerifyPassword 校验输入密码. 返回 (匹配, 是否需要升级存储格式).
func VerifyPassword(stored, input string) (ok bool, needUpgrade bool) {
	// 1. bcrypt 格式
	if strings.HasPrefix(stored, "$2") {
		err := bcrypt.CompareHashAndPassword([]byte(stored), []byte(input))
		return err == nil, false
	}
	// 2. PBKDF2 格式 "<hash>:<salt>"
	if strings.Contains(stored, ":") {
		parts := strings.SplitN(stored, ":", 2)
		if len(parts) != 2 {
			return false, false
		}
		salt, err := hex.DecodeString(parts[1])
		if err != nil {
			return false, false
		}
		derived := pbkdf2Encode(input, salt)
		match := subtle.ConstantTimeCompare([]byte(parts[0]), []byte(derived)) == 1
		return match, match // 匹配则升级到 bcrypt
	}
	// 3. 明文
	match := subtle.ConstantTimeCompare([]byte(stored), []byte(input)) == 1
	return match, match // 匹配则升级到 bcrypt
}

// genSaltUnused 保留旧 salt 生成器, 防止外部还有调用方 (实际未在新代码使用)
func genSaltUnused() ([]byte, error) {
	salt := make([]byte, passwordSaltLen)
	_, err := rand.Read(salt)
	return salt, err
}

// 生成token; role 进 claims 并被 HMAC 签名保护, 服务端鉴权一律读这里, 不读响应 body / 请求参数
func GenerateToken(secretKey string, iat, seconds int64, userID uint, username string, role int) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	claims["user_id"] = userID
	claims["username"] = username
	claims["role"] = role
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secretKey))
}

// 角色常量 (重复于 bean.User 是为了避免循环依赖)
const (
	RoleNormal = 0
	RoleAdmin  = 1
)

type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Role     int    `json:"role"`
}

// IsAdmin 是否超管
func (u *UserInfo) IsAdmin() bool {
	return u != nil && u.Role == RoleAdmin
}

// 鉴权 helper: 不是超管返回 error, 上层直接 return
var ErrPermissionDenied = fmt.Errorf("permission denied: admin required")

func RequireAdmin(ui *UserInfo) error {
	if !ui.IsAdmin() {
		return ErrPermissionDenied
	}
	return nil
}

func GetUserInfoFromCtx(ctx context.Context) (user *UserInfo, err error) {
	num, ok := ctx.Value("user_id").(json.Number)
	if !ok {
		return nil, fmt.Errorf("get user_id from ctx failed")
	}
	userId, err := num.Int64()
	if err != nil {
		return nil, fmt.Errorf("get user_id from ctx failed: num to int64 failed: %w", err)
	}
	username, ok := ctx.Value("username").(string)
	if !ok {
		return nil, fmt.Errorf("get username from ctx failed")
	}
	// role 可能缺失 (老 JWT 未含此 claim), 默认普通
	role := 0
	if v := ctx.Value("role"); v != nil {
		if n, ok2 := v.(json.Number); ok2 {
			if r, e := n.Int64(); e == nil {
				role = int(r)
			}
		}
	}
	return &UserInfo{
		ID:       uint(userId),
		Username: username,
		Role:     role,
	}, nil
}
