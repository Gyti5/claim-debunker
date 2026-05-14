package server

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type clientRateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	clients map[string]clientCounter
}

type clientCounter struct {
	count   int
	resetAt time.Time
}

func newClientRateLimiter(limit int, window time.Duration) *clientRateLimiter {
	return &clientRateLimiter{
		limit:   limit,
		window:  window,
		clients: make(map[string]clientCounter),
	}
}

func (l *clientRateLimiter) allow(clientID string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	counter := l.clients[clientID]
	if counter.resetAt.IsZero() || now.After(counter.resetAt) {
		l.clients[clientID] = clientCounter{
			count:   1,
			resetAt: now.Add(l.window),
		}
		l.cleanup(now)
		return true
	}
	if counter.count >= l.limit {
		return false
	}

	counter.count++
	l.clients[clientID] = counter
	return true
}

func (l *clientRateLimiter) cleanup(now time.Time) {
	if len(l.clients) < 1000 {
		return
	}

	for clientID, counter := range l.clients {
		if now.After(counter.resetAt) {
			delete(l.clients, clientID)
		}
	}
}

func withAPIKeyAuth(apiKey string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimSpace(apiKey) == "" {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "composition api key is not configured",
			})
			return
		}

		token := strings.TrimSpace(r.Header.Get("X-API-Key"))
		if token == "" {
			authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			}
		}

		if token == "" || token != apiKey {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "unauthorized",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func withRateLimit(limiter *clientRateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientID := extractClientID(r)
		if !limiter.allow(clientID, time.Now().UTC()) {
			writeJSON(w, http.StatusTooManyRequests, map[string]string{
				"error": "rate limit exceeded",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func extractClientID(r *http.Request) string {
	remoteAddr := strings.TrimSpace(r.RemoteAddr)
	if remoteAddr == "" {
		return "unknown"
	}

	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil && host != "" {
		return host
	}

	return remoteAddr
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
