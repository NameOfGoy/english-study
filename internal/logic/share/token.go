package share

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	TokenVersion = "v1"
	TokenTTL     = 5 * time.Minute

	ShareTypeAll   = 0 // 全部
	ShareTypeByTag = 1 // 按标签

	WordTypeAll    = 0
	WordTypeWord   = 1
	WordTypePhrase = 2
)

// SharePayload 是分享 token 的数据载荷
type SharePayload struct {
	UserID    uint32
	ShareType uint8
	WordType  uint8
	TagIDs    []uint32 // share_type=1 时使用
	ExpireTs  uint32   // unix seconds
	Nonce     uint64   // 64 位防重放熵 (原 uint16 只有 65536 值, 实际起不到防重放作用)
}

// IsExpired 判断 token 是否过期
func (p *SharePayload) IsExpired() bool {
	return time.Now().Unix() > int64(p.ExpireTs)
}

// EncodeToken 用 HMAC-SHA256 签名后编码成 "v1.<hex(payload)>.<hex(sig)>" 格式
func EncodeToken(payload *SharePayload, secret string) (string, error) {
	raw, err := marshalPayload(payload)
	if err != nil {
		return "", err
	}
	sig := signPayload(raw, secret)
	return fmt.Sprintf("%s.%s.%s", TokenVersion, hex.EncodeToString(raw), hex.EncodeToString(sig)), nil
}

// DecodeToken 验签并解码
func DecodeToken(token string, secret string) (*SharePayload, error) {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}
	if parts[0] != TokenVersion {
		return nil, fmt.Errorf("unsupported token version: %s", parts[0])
	}
	raw, err := hex.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode payload failed: %w", err)
	}
	gotSig, err := hex.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("decode signature failed: %w", err)
	}
	wantSig := signPayload(raw, secret)
	if !hmac.Equal(gotSig, wantSig) {
		return nil, errors.New("invalid signature")
	}
	payload, err := unmarshalPayload(raw)
	if err != nil {
		return nil, err
	}
	if payload.IsExpired() {
		return nil, errors.New("token expired")
	}
	return payload, nil
}

// NewNonce 生成 64 位随机 nonce
func NewNonce() uint64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return binary.BigEndian.Uint64(b[:])
}

func signPayload(raw []byte, secret string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(raw)
	// 用完整 32 字节 HMAC-SHA256 输出, 不再截断 (原来截到 16 字节让有效安全边际只剩 128 bit, 没必要)
	return mac.Sum(nil)
}

// DeriveShareSecret 从应用主密钥派生出分享 token 专用密钥（key separation）。
// 这样即使 share token 的密钥被攻破，也不会暴露 JWT 主密钥，
// 同时也防止 JWT 与 share token 互相通过验签。
const shareSecretLabel = "english-study/share-token/v1"

func DeriveShareSecret(masterSecret string) string {
	mac := hmac.New(sha256.New, []byte(masterSecret))
	mac.Write([]byte(shareSecretLabel))
	return hex.EncodeToString(mac.Sum(nil))
}

// 二进制布局:
//   user_id      uint32  (4)
//   share_type   uint8   (1)
//   word_type    uint8   (1)
//   expire_ts    uint32  (4)
//   nonce        uint64  (8)  ← 原 uint16, 不兼容老 token, 但 token TTL 仅 5min 无影响
//   tag_count    uint8   (1)
//   tag_ids      [tag_count]uint32
// 固定头 = 19 字节
const fixedHeaderSize = 19

func marshalPayload(p *SharePayload) ([]byte, error) {
	if len(p.TagIDs) > 255 {
		return nil, errors.New("too many tag_ids (max 255)")
	}
	size := fixedHeaderSize + 4*len(p.TagIDs)
	buf := make([]byte, 0, size)
	tmp := make([]byte, 8)

	binary.BigEndian.PutUint32(tmp[:4], p.UserID)
	buf = append(buf, tmp[:4]...)

	buf = append(buf, p.ShareType, p.WordType)

	binary.BigEndian.PutUint32(tmp[:4], p.ExpireTs)
	buf = append(buf, tmp[:4]...)

	binary.BigEndian.PutUint64(tmp, p.Nonce)
	buf = append(buf, tmp...)

	buf = append(buf, uint8(len(p.TagIDs)))
	for _, id := range p.TagIDs {
		binary.BigEndian.PutUint32(tmp[:4], id)
		buf = append(buf, tmp[:4]...)
	}
	return buf, nil
}

func unmarshalPayload(raw []byte) (*SharePayload, error) {
	if len(raw) < fixedHeaderSize {
		return nil, errors.New("payload too short")
	}
	p := &SharePayload{}
	p.UserID = binary.BigEndian.Uint32(raw[0:4])
	p.ShareType = raw[4]
	p.WordType = raw[5]
	p.ExpireTs = binary.BigEndian.Uint32(raw[6:10])
	p.Nonce = binary.BigEndian.Uint64(raw[10:18])
	tagCount := int(raw[18])
	expectedLen := fixedHeaderSize + 4*tagCount
	if len(raw) != expectedLen {
		return nil, fmt.Errorf("payload length mismatch: got %d, want %d", len(raw), expectedLen)
	}
	p.TagIDs = make([]uint32, tagCount)
	for i := 0; i < tagCount; i++ {
		p.TagIDs[i] = binary.BigEndian.Uint32(raw[fixedHeaderSize+4*i : fixedHeaderSize+4+4*i])
	}
	return p, nil
}
