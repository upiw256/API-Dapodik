package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"dapodik-readonly-api/internal/response"
)

type requestWindow struct {
	requests []time.Time
}

type rateLimiter struct {
	mu           sync.Mutex
	clients      map[string]requestWindow
	maxRequests  int
	window       time.Duration
	idleReset    time.Duration
	now          func() time.Time
	cleanupAfter time.Time
}

func RateLimit(maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	limiter := &rateLimiter{
		clients:     make(map[string]requestWindow),
		maxRequests: maxRequests,
		window:      window,
		idleReset:   time.Minute,
		now:         time.Now,
	}
	return limiter.middleware
}

func (l *rateLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions || l.maxRequests < 1 || l.window <= 0 {
			next.ServeHTTP(w, r)
			return
		}

		clientIP := remoteIP(r.RemoteAddr)
		now := l.now()
		l.mu.Lock()
		l.cleanupExpired(now)
		entry := l.clients[clientIP]
		cutoff := now.Add(-l.window)
		active := entry.requests[:0]
		for _, timestamp := range entry.requests {
			if timestamp.After(cutoff) {
				active = append(active, timestamp)
			}
		}
		entry.requests = active
		if len(entry.requests) > 0 && now.Sub(entry.requests[len(entry.requests)-1]) >= l.idleReset {
			entry.requests = entry.requests[:0]
		}
		if len(entry.requests) >= l.maxRequests {
			retryAfter := entry.requests[len(entry.requests)-1].Add(l.idleReset).Sub(now)
			seconds := int64((retryAfter + time.Second - 1) / time.Second)
			if seconds < 1 {
				seconds = 1
			}
			l.clients[clientIP] = entry
			l.mu.Unlock()
			w.Header().Set("Retry-After", strconv.FormatInt(seconds, 10))
			response.WriteError(w, http.StatusTooManyRequests, "rate limit exceeded; try again later")
			return
		}
		entry.requests = append(entry.requests, now)
		l.clients[clientIP] = entry
		l.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

func (l *rateLimiter) cleanupExpired(now time.Time) {
	if now.Before(l.cleanupAfter) {
		return
	}
	cutoff := now.Add(-l.window)
	for ip, entry := range l.clients {
		if len(entry.requests) == 0 || !entry.requests[len(entry.requests)-1].After(cutoff) {
			delete(l.clients, ip)
		}
	}
	l.cleanupAfter = now.Add(time.Minute)
}

func remoteIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		return host
	}
	return strings.Trim(remoteAddr, "[]")
}
