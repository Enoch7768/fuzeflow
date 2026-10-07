package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Enoch7768/fuzeflow/internal/auth"
	"github.com/Enoch7768/fuzeflow/internal/config"
	"github.com/Enoch7768/fuzeflow/internal/security"
)

type githubOAuth struct {
	store    *auth.Store
	secure   bool
	clientID string
	secret   string
	callback string
	client   *http.Client
}

func (o *githubOAuth) start(w http.ResponseWriter, r *http.Request) {
	if o.clientID == "" || o.secret == "" || o.callback == "" {
		http.Error(w, "GitHub sign-in is not configured", http.StatusNotImplemented)
		return
	}
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		http.Error(w, "unable to initialize sign-in", http.StatusInternalServerError)
		return
	}
	state := base64.RawURLEncoding.EncodeToString(stateBytes)
	http.SetCookie(w, &http.Cookie{
		Name: "ff_oauth_state", Value: state, Path: "/",
		MaxAge: 600, HttpOnly: true, Secure: o.secure, SameSite: http.SameSiteLaxMode,
	})
	query := url.Values{
		"client_id":    {o.clientID},
		"redirect_uri": {o.callback},
		"scope":        {"read:user user:email"},
		"state":        {state},
	}
	http.Redirect(w, r, "https://github.com/login/oauth/authorize?"+query.Encode(), http.StatusFound)
}

func (o *githubOAuth) callbackHandler(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie("ff_oauth_state")
	if err != nil || stateCookie.Value == "" || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "invalid OAuth state", http.StatusBadRequest)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "OAuth authorization failed", http.StatusBadRequest)
		return
	}

	token, err := o.exchangeCode(r, code)
	if err != nil {
		http.Error(w, "OAuth authorization failed", http.StatusUnauthorized)
		return
	}
	body, err := o.githubRequest(r, token, "https://api.github.com/user")
	if err != nil {
		http.Error(w, "unable to load GitHub identity", http.StatusBadGateway)
		return
	}

	var profile struct {
		ID    int64
		Login string
		Name  string
		Email string
	}
	if err := json.Unmarshal(body, &profile); err != nil || profile.ID == 0 {
		http.Error(w, "invalid GitHub identity", http.StatusBadGateway)
		return
	}

	email := strings.TrimSpace(profile.Email)
	if email == "" {
		email, err = o.githubEmail(r, token)
		if err != nil {
			http.Error(w, "a verified GitHub email is required", http.StatusBadRequest)
			return
		}
	}
	displayName := strings.TrimSpace(profile.Name)
	if displayName == "" {
		displayName = profile.Login
	}

	user, err := o.store.FindOrCreateOAuthUser(r.Context(), "github", fmt.Sprint(profile.ID), strings.ToLower(email), displayName)
	if err != nil {
		http.Error(w, "unable to create account", http.StatusInternalServerError)
		return
	}
	session, err := o.store.CreateSession(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "unable to create session", http.StatusInternalServerError)
		return
	}
	auth.SetSessionCookie(w, session, o.secure)
	_ = o.store.RecordAudit(r.Context(), nil, &user.ID, "auth.github_login", "user", user.ID.String(), security.RequestIDFromContext(r.Context()), r)
	http.SetCookie(w, &http.Cookie{Name: "ff_oauth_state", Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: o.secure, SameSite: http.SameSiteLaxMode})
	http.Redirect(w, r, "/", http.StatusFound)
}

func (o *githubOAuth) exchangeCode(r *http.Request, code string) (string, error) {
	form := url.Values{"client_id": {o.clientID}, "client_secret": {o.secret}, "code": {code}, "redirect_uri": {o.callback}}
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, "https://github.com/login/oauth/access_token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := o.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github token exchange status %d", resp.StatusCode)
	}
	var out map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil || out["access_token"] == "" {
		return "", fmt.Errorf("github token missing")
	}
	return out["access_token"], nil
}

func (o *githubOAuth) githubRequest(r *http.Request, token, endpoint string) ([]byte, error) {
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API status %d", resp.StatusCode)
	}
	var body json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	return body, nil
}

func (o *githubOAuth) githubEmail(r *http.Request, token string) (string, error) {
	body, err := o.githubRequest(r, token, "https://api.github.com/user/emails")
	if err != nil {
		return "", err
	}
	var emails []struct {
		Email    string
		Verified bool
		Primary   bool
	}
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", err
	}
	for _, email := range emails {
		if email.Verified && email.Primary {
			return email.Email, nil
		}
	}
	for _, email := range emails {
		if email.Verified {
			return email.Email, nil
		}
	}
	return "", fmt.Errorf("no verified email")
}

func newGitHubOAuth(cfg config.Config, store *auth.Store) *githubOAuth {
	return &githubOAuth{
		store: store, secure: cfg.CookieSecure,
		clientID: cfg.GitHubClientID, secret: cfg.GitHubClientSecret, callback: cfg.GitHubCallbackURL,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}
