package ccauth

import (
	"testing"
	"time"
)

func TestIPRateLimiter_LocksAfterMaxFailures(t *testing.T) {
	l := NewIPRateLimiter(3, 60, 60) // 3 次 / 60s 窗口, 锁 60s
	const ip = "1.2.3.4"

	for i := 0; i < 3; i++ {
		ok, _ := l.Allowed(ip)
		if !ok {
			t.Fatalf("attempt %d should be allowed", i)
		}
		l.RecordFailure(ip)
	}
	ok, rem := l.Allowed(ip)
	if ok {
		t.Fatal("expected locked after 3 failures")
	}
	if rem <= 0 || rem > 60 {
		t.Errorf("remaining %d should be 1-60", rem)
	}
}

func TestIPRateLimiter_SuccessClears(t *testing.T) {
	l := NewIPRateLimiter(3, 60, 60)
	const ip = "1.2.3.4"
	l.RecordFailure(ip)
	l.RecordFailure(ip) // 2 次失败
	l.RecordSuccess(ip)
	// 成功后计数器清零, 还能再连失败 3 次才锁
	l.RecordFailure(ip)
	l.RecordFailure(ip)
	if ok, _ := l.Allowed(ip); !ok {
		t.Fatal("should still be allowed after 2 fresh failures")
	}
}

func TestIPRateLimiter_DifferentIPsIndependent(t *testing.T) {
	l := NewIPRateLimiter(3, 60, 60)
	l.RecordFailure("a")
	l.RecordFailure("a")
	l.RecordFailure("a")
	if ok, _ := l.Allowed("b"); !ok {
		t.Error("IP b should not be locked by IP a failures")
	}
}

func TestRefreshStore_AddIsValidRevoke(t *testing.T) {
	s := NewRefreshStore()
	const jti = "abc"
	if s.IsValid(jti) {
		t.Fatal("unknown jti should be invalid")
	}
	s.Add(jti, time.Now().Unix()+60)
	if !s.IsValid(jti) {
		t.Fatal("freshly added jti should be valid")
	}
	s.Revoke(jti)
	if s.IsValid(jti) {
		t.Fatal("revoked jti should be invalid")
	}
}

func TestRefreshStore_ExpiredLazyEviction(t *testing.T) {
	s := NewRefreshStore()
	s.Add("x", time.Now().Unix()-1) // 已过期
	if s.IsValid("x") {
		t.Fatal("expired jti should report invalid")
	}
}
