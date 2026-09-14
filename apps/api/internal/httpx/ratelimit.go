package httpx

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type rateEntry struct {
	Count int
	Reset time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]rateEntry
	limit   int
	window  time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{entries: map[string]rateEntry{}, limit: limit, window: window}
}

func (l *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		now := time.Now()
		l.mu.Lock()
		e := l.entries[host]
		if e.Reset.Before(now) {
			e = rateEntry{Reset: now.Add(l.window)}
		}
		if e.Count >= l.limit {
			l.mu.Unlock()
			w.Header().Set("Retry-After", "60")
			Error(w, http.StatusTooManyRequests, "too many attempts, try again later")
			return
		}
		e.Count++
		l.entries[host] = e
		if len(l.entries) > 10000 {
			for k, v := range l.entries {
				if v.Reset.Before(now) {
					delete(l.entries, k)
				}
			}
		}
		l.mu.Unlock()
		next.ServeHTTP(w, r)
	})
}
