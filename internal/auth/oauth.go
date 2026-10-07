package auth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

func (s *Store) FindOrCreateOAuthUser(ctx context.Context, provider, providerUserID, email, displayName string) (User, error) {
	var user User
	err := s.db.QueryRowContext(ctx,
		"SELECT u.id, u.email, u.display_name, u.password_hash FROM oauth_accounts a JOIN users u ON u.id = a.user_id WHERE a.provider = $1 AND a.provider_user_id = $2",
		provider, providerUserID,
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.PasswordHash)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return User{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback()

	var existingID uuid.UUID
	err = tx.QueryRowContext(ctx, "SELECT id FROM users WHERE email = $1", email).Scan(&existingID)
	if err == nil {
		user = User{ID: existingID, Email: email, DisplayName: displayName}
	} else if errors.Is(err, sql.ErrNoRows) {
		existingID = uuid.New()
		if _, err := tx.ExecContext(ctx, "INSERT INTO users (id, email, display_name) VALUES ($1, $2, $3)", existingID, email, displayName); err != nil {
			return User{}, err
		}
		user = User{ID: existingID, Email: email, DisplayName: displayName}
	} else {
		return User{}, err
	}

	if _, err := tx.ExecContext(ctx, "INSERT INTO oauth_accounts (provider, provider_user_id, user_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING", provider, providerUserID, user.ID); err != nil {
		return User{}, err
	}
	if err := tx.Commit(); err != nil {
		return User{}, err
	}
	return user, nil
}
