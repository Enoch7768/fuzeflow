package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/Enoch7768/fuzeflow/internal/auth"
	"github.com/Enoch7768/fuzeflow/internal/security"
)

type Handler struct {
	logger *slog.Logger
}

func NewHandler(logger *slog.Logger) http.Handler {
	return newHandler(logger, nil, false)
}

func NewHandlerWithStore(logger *slog.Logger, store *auth.Store, secureCookies bool) http.Handler {
	return newHandler(logger, store, secureCookies)
}

func newHandler(logger *slog.Logger, store *auth.Store, secureCookies bool) http.Handler {
	h := &Handler{logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("GET /readyz", h.ready)
	mux.HandleFunc("GET /api/v1", h.apiRoot)

	if store != nil {
		authAPI := &authAPI{store: store, secure: secureCookies}
		limited := security.NewLimiter(10, time.Minute)
		mux.Handle("POST /api/v1/auth/signup", rateLimited(limited, http.HandlerFunc(authAPI.signup)))
		mux.Handle("POST /api/v1/auth/login", rateLimited(limited, http.HandlerFunc(authAPI.login)))
		mux.HandleFunc("POST /api/v1/auth/logout", auth.WithSession(store, http.HandlerFunc(authAPI.logout)))
		mux.Handle("GET /api/v1/auth/me", auth.WithSession(store, auth.RequireAuth(http.HandlerFunc(authAPI.me))))
		mux.Handle("GET /api/v1/organizations", auth.WithSession(store, auth.RequireAuth(http.HandlerFunc(authAPI.organizations))))
	}

	handler := security.RequestID(security.NoStore(security.CleanupLimiter(securityHeaders(requestLogging(mux, logger)))))
	return handler
}

func rateLimited(limiter *security.Limiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow(r) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) ready(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (h *Handler) apiRoot(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"name":    "FuzeFlow",
		"version": "0.1.0",
		"status":  "security-foundation",
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func requestLogging(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("http request", "method", r.Method, "path", r.URL.Path, "duration_ms", time.Since(start).Milliseconds())
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, r)
	})
}
