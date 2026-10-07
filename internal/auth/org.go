package auth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

var ErrForbidden = errors.New("organization access denied")

type Organization struct {
	ID   uuid.UUID
	Name string
	Role string
}

func (s *Store) CreateOrganization(ctx context.Context, userID uuid.UUID, name string) (Organization, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Organization{}, err
	}
	defer tx.Rollback()

	orgID := uuid.New()
	if _, err := tx.ExecContext(ctx, "INSERT INTO organizations (id, name) VALUES ($1, $2)", orgID, name); err != nil {
		return Organization{}, err
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO organization_members (organization_id, user_id, role) VALUES ($1, $2, 'owner')", orgID, userID); err != nil {
		return Organization{}, err
	}
	if err := tx.Commit(); err != nil {
		return Organization{}, err
	}
	return Organization{ID: orgID, Name: name, Role: "owner"}, nil
}

func (s *Store) OrganizationsForUser(ctx context.Context, userID uuid.UUID) ([]Organization, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT o.id, o.name, m.role FROM organizations o JOIN organization_members m ON m.organization_id = o.id WHERE m.user_id = $1 ORDER BY o.name",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var organizations []Organization
	for rows.Next() {
		var org Organization
		if err := rows.Scan(&org.ID, &org.Name, &org.Role); err != nil {
			return nil, err
		}
		organizations = append(organizations, org)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return organizations, nil
}

func (s *Store) OrganizationForUser(ctx context.Context, userID, organizationID uuid.UUID) (Organization, error) {
	var org Organization
	err := s.db.QueryRowContext(ctx,
		"SELECT o.id, o.name, m.role FROM organizations o JOIN organization_members m ON m.organization_id = o.id WHERE o.id = $1 AND m.user_id = $2",
		organizationID, userID,
	).Scan(&org.ID, &org.Name, &org.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return Organization{}, ErrForbidden
	}
	return org, err
}
