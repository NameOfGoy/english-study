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
	// passwordSymbols 密码允许且计入"符号"类的安全符号白名单(明确不含 引号/反斜杠/空格/尖括号), 与前端同一份字符集
	passwordSymbols = "!@#$%^&*()-_=+[]{};:,.?"
)

// ValidatePasswordStrength 校验密码强度(注册/改密用; 登录不调用, 否则历史弱密码会被锁死).
// 规则: 长度 8~64; 仅允许 字母/数字/安全符号; 大写/小写/数字/符号 四类中至少含 3 类("4含3").
func ValidatePasswordStrength(pwd string) error {
	if len(pwd) < 8 {
		return fmt.Errorf("密码至少 8 位")
	}
	if len(pwd) > 64 {
		return fmt.Errorf("密码不超过 64 位")
	}
	var hasUpper, hasLower, hasDigit, hasSymbol bool
	for _, r := range pwd {
		switch {
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= '0' && r <= '9':
			hasDigit = true
		case strings.ContainsRune(passwordSymbols, r):
			hasSymbol = true
		default:
			return fmt.Errorf("密码含不支持的字符, 仅允许字母/数字及 %s", passwordSymbols)
		}
	}
	cats := 0
	for _, ok := range []bool{hasUpper, hasLower, hasDigit, hasSymbol} {
		if ok {
			cats++
		}
	}
	if cats < 3 {
		return fmt.Errorf("密码需包含大写/小写/数字/符号中至少 3 类")
	}
	return nil
}

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

// RelayAudience 转发器 (cc 桥) 专用 JWT aud claim.
// 加 aud 是为了防 english-study 业务 token 被滥用接入 forwarder ws —— 即便 secret 共享,
// 业务 token 没 aud=forwarder, 转发器会 401 拒.
const RelayAudience = "forwarder"

// GenerateRelayToken 旧接口, 保留兼容; 新代码请用 GenerateRelayAccessToken.
// 默认签发 access token (token_type=access), 不带 jti, 不可 rotation.
func GenerateRelayToken(secretKey string, iat, ttlSec int64, userID uint, username string, role int) (string, error) {
	return GenerateRelayAccessToken(secretKey, iat, ttlSec, userID, username, role, "")
}

// Token 类型常量, 写进 JWT claim "token_type", 转发器 + 后端 /relay-refresh 据此区分接受范围.
const (
	RelayTokenTypeAccess  = "access"
	RelayTokenTypeRefresh = "refresh"
)

// GenerateRelayAccessToken 签发短 TTL access token. 用于 ws 鉴权.
//   jti: 可空; 非空时写入 claim, 配合 refresh rotation 关联两条 token 寿命
func GenerateRelayAccessToken(secretKey string, iat, ttlSec int64, userID uint, username string, role int, jti string) (string, error) {
	claims := jwt.MapClaims{
		"exp":        iat + ttlSec,
		"iat":        iat,
		"user_id":    userID,
		"username":   username,
		"role":       role,
		"aud":        RelayAudience,
		"token_type": RelayTokenTypeAccess,
	}
	if jti != "" {
		claims["jti"] = jti
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// GenerateRelayRefreshToken 签发长 TTL refresh token. 仅 /api/v1/cc/relay-refresh 接受;
// 转发器 ws 显式拒绝 token_type=refresh, 防止 refresh 被当 access 用.
//   jti 必填: 唯一标识本条 refresh, /relay-refresh 用它做 rotation 后撤旧
func GenerateRelayRefreshToken(secretKey string, iat, ttlSec int64, userID uint, username string, role int, jti string) (string, error) {
	if jti == "" {
		return "", fmt.Errorf("refresh token must have jti for rotation tracking")
	}
	claims := jwt.MapClaims{
		"exp":        iat + ttlSec,
		"iat":        iat,
		"user_id":    userID,
		"username":   username,
		"role":       role,
		"aud":        RelayAudience,
		"token_type": RelayTokenTypeRefresh,
		"jti":        jti,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// ===== AI 桥文件下载签名 token =====
//
// 私有桶 cc-uploads 不可匿名读. CC(本地) 要下载文件, 既不该拿业务 JWT(隐式耦合),
// 也连不到内网 MinIO 做原生 presigned. 方案: 后端用同一 HS256 secret 签一个短时效
// token (内含 bucket|object + exp), CC 凭 token 走 GET /api/v1/cc/download 流式下载 —
// 等价于"应用层 presigned": 自鉴权 + 带 TTL + 不暴露 MinIO.

const RelayFileAudience = "cc-file"

// GenerateFileToken 为 (bucket, object) 签一个下载 token. ttlSec 后过期.
func GenerateFileToken(secretKey, bucket, object string, iat, ttlSec int64) (string, error) {
	claims := jwt.MapClaims{
		"exp": iat + ttlSec,
		"iat": iat,
		"aud": RelayFileAudience,
		// 用 "|" 分隔 bucket 与 object: object 自身含 "/", 用 "|" 不歧义
		"obj": bucket + "|" + object,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// VerifyFileToken 校验下载 token, 返回 (bucket, object). 校验签名 + 过期 + aud=cc-file.
func VerifyFileToken(secretKey, token string) (bucket, object string, err error) {
	var claims jwt.MapClaims
	t, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected alg: %v", t.Method.Alg())
		}
		return []byte(secretKey), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return "", "", err
	}
	if !t.Valid {
		return "", "", fmt.Errorf("file token invalid")
	}
	if aud, _ := claims["aud"].(string); aud != RelayFileAudience {
		// 防止把业务 token / relay token 拿来当下载 token 用
		return "", "", fmt.Errorf("file token aud mismatch")
	}
	obj, _ := claims["obj"].(string)
	parts := strings.SplitN(obj, "|", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("file token obj malformed")
	}
	return parts[0], parts[1], nil
}

// 角色常量 (重复于 bean.User 是为了避免循环依赖)
const (
	RoleNormal = 0
	RoleAdmin  = 1
	// RoleGuest 游客: 微信登录但 openid 未注册时后端发的只读身份。
	// 守卫中间件只放行游客的 GET(只读浏览), 写操作(非 GET)一律拦下引导登录。
	RoleGuest = 2
)

// GuestAccountName 固定游客账号的 account 值。需预先建好该账号并 seed 示例数据(词/短语/文章),
// 微信登录未命中真实账号时, 后端签发这个账号的只读 token, 供游客先浏览体验(过审"先体验后授权")。
const GuestAccountName = "wxappguest"

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
