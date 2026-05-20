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
	Nonce     uint16
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

// NewNonce 生成随机 nonce
func NewNonce() uint16 {
	var b [2]byte
	_, _ = rand.Read(b[:])
	return binary.BigEndian.Uint16(b[:])
}

func signPayload(raw []byte, secret string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(raw)
	// truncate to 16 bytes 已经足够防伪造
	return mac.Sum(nil)[:16]
}

// 二进制布局:
//   user_id      uint32
//   share_type   uint8
//   word_type    uint8
//   expire_ts    uint32
//   nonce        uint16
//   tag_count    uint8
//   tag_ids      [tag_count]uint32
func marshalPayload(p *SharePayload) ([]byte, error) {
	if len(p.TagIDs) > 255 {
		return nil, errors.New("too many tag_ids (max 255)")
	}
	size := 4 + 1 + 1 + 4 + 2 + 1 + 4*len(p.TagIDs)
	buf := make([]byte, 0, size)
	tmp := make([]byte, 4)

	binary.BigEndian.PutUint32(tmp, p.UserID)
	buf = append(buf, tmp...)

	buf = append(buf, p.ShareType, p.WordType)

	binary.BigEndian.PutUint32(tmp, p.ExpireTs)
	buf = append(buf, tmp...)

	binary.BigEndian.PutUint16(tmp[:2], p.Nonce)
	buf = append(buf, tmp[:2]...)

	buf = append(buf, uint8(len(p.TagIDs)))
	for _, id := range p.TagIDs {
		binary.BigEndian.PutUint32(tmp, id)
		buf = append(buf, tmp...)
	}
	return buf, nil
}

func unmarshalPayload(raw []byte) (*SharePayload, error) {
	if len(raw) < 13 {
		return nil, errors.New("payload too short")
	}
	p := &SharePayload{}
	p.UserID = binary.BigEndian.Uint32(raw[0:4])
	p.ShareType = raw[4]
	p.WordType = raw[5]
	p.ExpireTs = binary.BigEndian.Uint32(raw[6:10])
	p.Nonce = binary.BigEndian.Uint16(raw[10:12])
	tagCount := int(raw[12])
	expectedLen := 13 + 4*tagCount
	if len(raw) != expectedLen {
		return nil, fmt.Errorf("payload length mismatch: got %d, want %d", len(raw), expectedLen)
	}
	p.TagIDs = make([]uint32, tagCount)
	for i := 0; i < tagCount; i++ {
		p.TagIDs[i] = binary.BigEndian.Uint32(raw[13+4*i : 17+4*i])
	}
	return p, nil
}
