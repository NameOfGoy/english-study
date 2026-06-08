// Package ccauth AI 桥二次门禁的进程内状态: IP 错密钥锁 + refresh jti rotation store.
//
// 单租户 / 单实例够用; 多实例部署要换 Redis (改 store 接口实现就行).
// 这里独立成包是为了避免 svc 反向 import logic/cc 形成循环.
package ccauth

import (
	"context"
	"sync"
	"time"
)

// ============= IP 错密钥锁定 =============

// IPRateLimiter 滑窗错密钥锁: 同一 IP 在 windowSec 内错 maxFailures 次 → lockSec 内全拒.
type IPRateLimiter struct {
	mu          sync.Mutex
	maxFailures int
	windowSec   int64
	lockSec     int64
	records     map[string]*ipRecord
}

type ipRecord struct {
	failures    []int64
	lockedUntil int64
}

func NewIPRateLimiter(maxFailures int, windowSec, lockSec int64) *IPRateLimiter {
	return &IPRateLimiter{
		maxFailures: maxFailures,
		windowSec:   windowSec,
		lockSec:     lockSec,
		records:     map[string]*ipRecord{},
	}
}

// Allowed 返回 (true, 0) 表示可以尝试; (false, 还需多少秒解锁).
func (l *IPRateLimiter) Allowed(ip string) (bool, int64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now().Unix()
	r := l.records[ip]
	if r == nil {
		return true, 0
	}
	if r.lockedUntil > now {
		return false, r.lockedUntil - now
	}
	return true, 0
}

func (l *IPRateLimiter) RecordFailure(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now().Unix()
	r := l.records[ip]
	if r == nil {
		r = &ipRecord{}
		l.records[ip] = r
	}
	cutoff := now - l.windowSec
	kept := r.failures[:0]
	for _, t := range r.failures {
		if t > cutoff {
			kept = append(kept, t)
		}
	}
	r.failures = append(kept, now)
	if len(r.failures) >= l.maxFailures {
		r.lockedUntil = now + l.lockSec
		r.failures = nil
	}
}

func (l *IPRateLimiter) RecordSuccess(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.records, ip)
}

// ============= refresh token jti rotation store =============

// RefreshStore 保存"有效" refresh jti 集合; rotation 时旧 jti 撤销, 新 jti 入.
type RefreshStore struct {
	mu    sync.Mutex
	valid map[string]int64 // jti → 过期时间 unix sec
}

func NewRefreshStore() *RefreshStore {
	s := &RefreshStore{valid: map[string]int64{}}
	go s.gcLoop()
	return s
}

func (s *RefreshStore) gcLoop() {
	t := time.NewTicker(10 * time.Minute)
	defer t.Stop()
	for range t.C {
		s.gc()
	}
}

func (s *RefreshStore) gc() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().Unix()
	for jti, exp := range s.valid {
		if exp <= now {
			delete(s.valid, jti)
		}
	}
}

func (s *RefreshStore) Add(jti string, expiresAt int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.valid[jti] = expiresAt
}

func (s *RefreshStore) IsValid(jti string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	exp, ok := s.valid[jti]
	if !ok {
		return false
	}
	if exp <= time.Now().Unix() {
		delete(s.valid, jti)
		return false
	}
	return true
}

func (s *RefreshStore) Revoke(jti string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.valid, jti)
}

// ============= client IP from ctx =============

type clientIPKey struct{}

// WithClientIP 把 client IP 塞进 ctx, 给 logic 取出做 IP 锁判断.
// handler 收到 *http.Request 后调用.
func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, clientIPKey{}, ip)
}

func ClientIPFromCtx(ctx context.Context) string {
	if v, ok := ctx.Value(clientIPKey{}).(string); ok {
		return v
	}
	return ""
}
