package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidSession = errors.New("invalid session")

const SessionTTL = 7 * 24 * time.Hour

type User struct {
	ID           uuid.UUID
	Email        string
	DisplayName  string
	PasswordHash string
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateUser(ctx context.Context, email, displayName, passwordHash string) (User, error) {
	id := uuid.New()
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO users (id, email, display_name, password_hash) VALUES ($1, $2, $3, $4)",
		id, email, displayName, passwordHash,
	)
	if err != nil {
		return User{}, err
	}
	return User{ID: id, Email: email, DisplayName: displayName, PasswordHash: passwordHash}, nil
}

func (s *Store) FindUserByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := s.db.QueryRowContext(ctx,
		"SELECT id, email, display_name, password_hash FROM users WHERE email = $1",
		email,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash)
	return u, err
}

func (s *Store) CreateSession(ctx context.Context, userID uuid.UUID) (string, error) {
	token, hash, err := NewSessionToken()
	if err != nil {
		return "", err
	}
	_, err = s.db.ExecContext(ctx,
		"INSERT INTO sessions (id, user_id, token_hash, expires_at) VALUES ($1, $2, $3, $4)",
		uuid.New(), userID, hash, time.Now().UTC().Add(SessionTTL),
	)
	return token, err
}

func (s *Store) FindUserBySession(ctx context.Context, token string) (User, error) {
	hash := HashSessionToken(token)
	if hash == nil {
		return User{}, ErrInvalidSession
	}
	var u User
	err := s.db.QueryRowContext(ctx,
		"SELECT u.id, u.email, u.display_name, u.password_hash FROM sessions s JOIN users u ON u.id = s.user_id WHERE s.token_hash = $1 AND s.expires_at > NOW()",
		hash,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrInvalidSession
	}
	if err != nil {
		return User{}, err
	}
	_, _ = s.db.ExecContext(ctx, "UPDATE sessions SET last_seen_at = NOW() WHERE token_hash = $1", hash)
	return u, nil
}

func (s *Store) RevokeSession(ctx context.Context, token string) error {
	hash := HashSessionToken(token)
	if hash == nil {
		return ErrInvalidSession
	}
	_, err := s.db.ExecContext(ctx, "DELETE FROM sessions WHERE token_hash = $1", hash)
	return err
}
