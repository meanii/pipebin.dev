package middleware

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/meanii/pipebin.dev/services/api/internal/httpx"
)

const (
	// ratePerMinute is the number of paste-creation requests allowed per IP per minute.
	ratePerMinute = 20
	// burstSize is the maximum burst above the steady rate.
	burstSize = 5
	// ttl is how long an idle IP entry is kept in memory.
	ttl = 10 * time.Minute
	// cleanupInterval is how often stale entries are pruned.
	cleanupInterval = 5 * time.Minute
)

type entry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	mu      sync.Mutex
	clients = make(map[string]*entry)
)

func init() {
	// Background goroutine that evicts entries that haven't been seen for ttl.
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			for ip, e := range clients {
				if time.Since(e.lastSeen) > ttl {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()
}

func getLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()
	e, ok := clients[ip]
	if !ok {
		e = &entry{limiter: rate.NewLimiter(rate.Every(time.Minute/ratePerMinute), burstSize)}
		clients[ip] = e
	}
	e.lastSeen = time.Now()
	return e.limiter
}

// RateLimitMiddleware enforces per-IP rate limiting on paste creation (POST /).
// All other routes are passed through untouched.
func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		ip := httpx.GetClientIP(r)
		if !getLimiter(ip).Allow() {
			slog.WarnContext(r.Context(), "rate limit exceeded",
				slog.String("path", r.URL.Path),
			)
			w.Header().Set("Retry-After", "60")
			httpx.EResponse(w, "rate limit exceeded — try again in a minute", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
