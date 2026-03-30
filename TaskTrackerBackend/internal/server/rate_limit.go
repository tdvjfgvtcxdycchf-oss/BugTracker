package server

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type rateLimiter struct {
	mu      sync.Mutex
	hits    map[string][]time.Time
	limit   int
	window  time.Duration
	cleanup time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		hits:    make(map[string][]time.Time),
		limit:   limit,
		window:  window,
		cleanup: 5 * time.Minute,
	}
	go rl.gcLoop()
	return rl
}

func (r *rateLimiter) allow(key string) bool {
	now := time.Now()
	cutoff := now.Add(-r.window)

	r.mu.Lock()
	defer r.mu.Unlock()

	entries := r.hits[key]
	dst := entries[:0]
	for _, t := range entries {
		if t.After(cutoff) {
			dst = append(dst, t)
		}
	}
	entries = dst
	if len(entries) >= r.limit {
		r.hits[key] = entries
		return false
	}
	entries = append(entries, now)
	r.hits[key] = entries
	return true
}

func (r *rateLimiter) gcLoop() {
	ticker := time.NewTicker(r.cleanup)
	defer ticker.Stop()
	for range ticker.C {
		r.mu.Lock()
		cutoff := time.Now().Add(-r.window)
		for k, entries := range r.hits {
			alive := false
			for _, t := range entries {
				if t.After(cutoff) {
					alive = true
					break
				}
			}
			if !alive {
				delete(r.hits, k)
			}
		}
		r.mu.Unlock()
	}
}

func clientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && ip != "" {
		return ip
	}
	return r.RemoteAddr
}

func withRateLimit(rl *rateLimiter, keyPrefix string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := keyPrefix + ":" + clientIP(r)
		if !rl.allow(key) {
			writeJSONErrorCode(w, http.StatusTooManyRequests, "rate_limited", "too many requests, try later")
			return
		}
		next.ServeHTTP(w, r)
	})
}

