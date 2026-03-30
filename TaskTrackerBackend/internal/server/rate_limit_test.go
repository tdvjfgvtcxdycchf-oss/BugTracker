package server

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl := newRateLimiter(5, time.Minute)
	for i := 0; i < 5; i++ {
		if !rl.allow("key") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
}

func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	rl := newRateLimiter(3, time.Minute)
	for i := 0; i < 3; i++ {
		rl.allow("key")
	}
	if rl.allow("key") {
		t.Fatal("4th request should be blocked")
	}
}

func TestRateLimiter_SeparateKeys(t *testing.T) {
	rl := newRateLimiter(2, time.Minute)
	rl.allow("ip1")
	rl.allow("ip1")
	// ip1 is now exhausted; ip2 should still be allowed
	if !rl.allow("ip2") {
		t.Fatal("ip2 should be allowed independently")
	}
	// ip1 should be blocked
	if rl.allow("ip1") {
		t.Fatal("ip1 should be blocked after 2 requests")
	}
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	rl := newRateLimiter(2, 50*time.Millisecond)
	rl.allow("key")
	rl.allow("key")
	if rl.allow("key") {
		t.Fatal("should be blocked at limit")
	}
	time.Sleep(60 * time.Millisecond)
	// Window expired — requests should be allowed again
	if !rl.allow("key") {
		t.Fatal("should be allowed after window expiry")
	}
}

func TestRateLimiter_Concurrent(t *testing.T) {
	limit := 10
	rl := newRateLimiter(limit, time.Minute)
	var allowed, blocked int
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ok := rl.allow("key")
			mu.Lock()
			if ok {
				allowed++
			} else {
				blocked++
			}
			mu.Unlock()
		}()
	}
	wg.Wait()
	if allowed != limit {
		t.Errorf("want %d allowed, got %d", limit, allowed)
	}
	if blocked != 20-limit {
		t.Errorf("want %d blocked, got %d", 20-limit, blocked)
	}
}

func TestClientIP_XForwardedFor(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8")
	r.RemoteAddr = "9.10.11.12:1234"
	got := clientIP(r)
	if got != "1.2.3.4" {
		t.Errorf("want 1.2.3.4, got %s", got)
	}
}

func TestClientIP_RemoteAddr(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.1:9999"
	got := clientIP(r)
	if got != "10.0.0.1" {
		t.Errorf("want 10.0.0.1, got %s", got)
	}
}

func TestWithRateLimit_Returns429WhenExhausted(t *testing.T) {
	rl := newRateLimiter(1, time.Minute)
	called := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		w.WriteHeader(http.StatusOK)
	})
	handler := withRateLimit(rl, "test", next)

	// First request — should pass through
	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("first request: want 200, got %d", rr.Code)
	}

	// Second request — should be rate limited
	req2 := httptest.NewRequest(http.MethodPost, "/login", nil)
	req2.RemoteAddr = "1.2.3.4:1234"
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("second request: want 429, got %d", rr2.Code)
	}
	if called != 1 {
		t.Errorf("next handler should be called exactly once, got %d", called)
	}
}
