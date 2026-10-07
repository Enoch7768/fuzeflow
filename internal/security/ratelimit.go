package security

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Limiter struct {
	mu      sync.Mutex
	clients map[string]*bucket
	limit   int
	window  time.Duration
}

type bucket struct {
	count int
	reset time.Time
}

func NewLimiter(limit int, window time.Duration) *Limiter {
	return &Limiter{clients: make(map[string]*bucket), limit: limit, window: window}
}

func (l *Limiter) Allow(r *http.Request) bool {
	key := clientKey(r)
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	b := l.clients[key]
	if b == nil || now.After(b.reset) {
		l.clients[key] = &bucket{count: 1, reset: now.Add(l.window)}
		return true
	}
	if b.count >= l.limit {
		return false
	}
	b.count++
	return true
}

func clientKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}
