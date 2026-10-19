package api

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"gooava/internal/models"
)

/************************************************************************
* rateLimiter is a fixed-window, per-client-IP rate limiter for the public
* auth endpoints (login, signup, email verification, password reset). These
* endpoints guard every account — passwords and 6-digit verification codes
* must be expensive to brute force.
*
* In-memory is fine here: the app runs as a single container on one VPS.
*
* status: ✅
************************************************************************/
type rateLimiter struct {
	mu        sync.Mutex
	visitors  map[string]*visitor
	max       int
	window    time.Duration
	lastSweep time.Time
}

type visitor struct {
	count       int
	windowStart time.Time
}

func newRateLimiter(max int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		visitors:  make(map[string]*visitor),
		max:       max,
		window:    window,
		lastSweep: time.Now(),
	}
}

/************************************************************************
* allow reports whether the given IP may make another request right now.
* Expired windows reset the counter; a periodic sweep drops stale entries
* so the map does not grow forever.
************************************************************************/
func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	if now.Sub(rl.lastSweep) > rl.window {
		for key, v := range rl.visitors {
			if now.Sub(v.windowStart) > rl.window {
				delete(rl.visitors, key)
			}
		}
		rl.lastSweep = now
	}

	v, ok := rl.visitors[ip]
	if !ok || now.Sub(v.windowStart) > rl.window {
		rl.visitors[ip] = &visitor{count: 1, windowStart: now}
		return true
	}

	if v.count >= rl.max {
		return false
	}

	v.count++
	return true
}

/************************************************************************
* clientIP resolves the real client IP. In production nginx proxies every
* request (RemoteAddr is always 127.0.0.1), so we trust X-Real-IP or the
* LAST X-Forwarded-For entry — the one our nginx appended; earlier entries
* are client-supplied and spoofable. Direct hits (local dev) fall back to
* RemoteAddr.
*
* NOTE: requires nginx to set `proxy_set_header X-Real-IP $remote_addr;`
************************************************************************/
func clientIP(r *http.Request) string {
	if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
		return xri
	}

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[len(parts)-1])
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

/************************************************************************
* wrap applies the limiter to a handler. Clients over the limit get
* HTTP 429 in the standard HttpResponse shape.
************************************************************************/
func (rl *rateLimiter) wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !rl.allow(clientIP(r)) {
			var response models.HttpResponse
			response.Success = false
			response.Error = "Too many attempts. Please wait a few minutes and try again."
			response.Data = nil

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			response.Send(w)
			return
		}

		next(w, r)
	}
}
