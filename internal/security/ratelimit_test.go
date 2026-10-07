package security

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestLimiter(t *testing.T) {
	l := NewLimiter(2, time.Minute)
	r := httptest.NewRequest("POST", "/login", nil)
	r.RemoteAddr = "127.0.0.1:1234"

	if !l.Allow(r) || !l.Allow(r) {
		t.Fatal("expected first two requests to pass")
	}
	if l.Allow(r) {
		t.Fatal("expected third request to be blocked")
	}
}
