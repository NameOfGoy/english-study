package utils

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/pbkdf2"
)

func PasswordEncode(password string, salt []byte) (key string) {
	// 生成密文
	dk := pbkdf2.Key([]byte(password), salt, 2048, 32, sha256.New)
	return hex.EncodeToString(dk)
}

// 生成token
func GenerateToken(secretKey string, iat, seconds int64, userID uint, username string) (string, error) {
	// 生成token
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	claims["user_id"] = userID
	claims["username"] = username
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secretKey))
}

type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
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
	return &UserInfo{
		ID:       uint(userId),
		Username: username,
	}, nil
}
