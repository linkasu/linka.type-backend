package coreapi

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter manages per-IP rate limiting
type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*visitorLimiter
	rate     rate.Limit
	burst    int
	ttl      time.Duration
}

type visitorLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter
// rate is requests per second, burst is max burst size
func NewRateLimiter(r float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*visitorLimiter),
		rate:     rate.Limit(r),
		burst:    burst,
		ttl:      3 * time.Minute,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.limiters[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[ip] = &visitorLimiter{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}
	v.lastSeen = time.Now()
	return v.limiter
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.limiters {
			if time.Since(v.lastSeen) > rl.ttl {
				delete(rl.limiters, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware returns HTTP middleware that limits requests per IP
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		limiter := rl.getLimiter(ip)
		if !limiter.Allow() {
			http.Error(w, `{"error":{"code":"rate_limited","message":"too many requests"}}`, http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies like Yandex Cloud)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP in the chain
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// Rate limiters for different endpoints
var (
	// AuthRateLimiter: 5 requests per minute for auth endpoints (brute-force protection)
	AuthRateLimiter = NewRateLimiter(5.0/60.0, 5)

	// APIRateLimiter: 100 requests per minute for general API
	APIRateLimiter = NewRateLimiter(100.0/60.0, 20)

	// PredictorRateLimiter: 50 requests per minute for predictor
	PredictorRateLimiter = NewRateLimiter(50.0/60.0, 10)
)
