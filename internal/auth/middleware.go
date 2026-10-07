package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type contextKey string

const userKey contextKey = "authenticated_user"
const SessionCookie = "ff_session"

func WithSession(store *Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(SessionCookie)
		if err == nil && cookie.Value != "" {
			if user, err := store.FindUserBySession(r.Context(), cookie.Value); err == nil {
				r = r.WithContext(context.WithValue(r.Context(), userKey, user))
			}
		}
		next.ServeHTTP(w, r)
	})
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := UserFromContext(r.Context()); !ok {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func UserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userKey).(User)
	return user, ok
}

func SetSessionCookie(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name: SessionCookie, Value: token, Path: "/",
		Expires: time.Now().UTC().Add(SessionTTL),
		MaxAge: int(SessionTTL / time.Second),
		HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode,
	})
}

func ClearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name: SessionCookie, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode,
	})
}

func UserID(ctx context.Context) (uuid.UUID, bool) {
	user, ok := UserFromContext(ctx)
	if !ok {
		return uuid.Nil, false
	}
	return user.ID, true
}
