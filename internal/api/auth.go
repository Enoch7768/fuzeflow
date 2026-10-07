package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Enoch7768/fuzeflow/internal/auth"
)

type authAPI struct {
	store  *auth.Store
	secure bool
}

type credentials struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

func (a *authAPI) signup(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if err := decodeJSON(w, r, &in); err != nil {
		return
	}
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	in.DisplayName = strings.TrimSpace(in.DisplayName)
	if in.Email == "" || in.DisplayName == "" || len(in.Password) < 12 {
		http.Error(w, "invalid registration data", http.StatusBadRequest)
		return
	}
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		http.Error(w, "unable to create account", http.StatusInternalServerError)
		return
	}
	user, err := a.store.CreateUser(r.Context(), in.Email, in.DisplayName, hash)
	if err != nil {
		http.Error(w, "unable to create account", http.StatusConflict)
		return
	}
	token, err := a.store.CreateSession(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "unable to create session", http.StatusInternalServerError)
		return
	}
	auth.SetSessionCookie(w, token, a.secure)
	writeJSON(w, http.StatusCreated, map[string]any{"id": user.ID, "email": user.Email, "display_name": user.DisplayName})
}

func (a *authAPI) login(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if err := decodeJSON(w, r, &in); err != nil {
		return
	}
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	user, err := a.store.FindUserByEmail(r.Context(), in.Email)
	if err != nil || user.PasswordHash == "" || !auth.CheckPassword(user.PasswordHash, in.Password) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	token, err := a.store.CreateSession(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "unable to create session", http.StatusInternalServerError)
		return
	}
	auth.SetSessionCookie(w, token, a.secure)
	writeJSON(w, http.StatusOK, map[string]any{"id": user.ID, "email": user.Email, "display_name": user.DisplayName})
}

func (a *authAPI) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(auth.SessionCookie); err == nil {
		_ = a.store.RevokeSession(r.Context(), cookie.Value)
	}
	auth.ClearSessionCookie(w, a.secure)
	w.WriteHeader(http.StatusNoContent)
}

func (a *authAPI) me(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	writeJSON(w, http.StatusOK, map[string]any{"id": user.ID, "email": user.Email, "display_name": user.DisplayName})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return err
	}
	return nil
}
