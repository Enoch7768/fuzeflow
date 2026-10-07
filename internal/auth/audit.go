package auth

import (
	"context"
	"encoding/json"
	"net"
	"net/http"

	"github.com/google/uuid"
)

func (s *Store) RecordAudit(ctx context.Context, organizationID, userID *uuid.UUID, action, resourceType, resourceID, requestID string, r *http.Request) error {
	var ip any
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		ip = host
	}
	metadata := json.RawMessage("{}")
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO audit_logs (organization_id, user_id, action, resource_type, resource_id, request_id, ip_address, metadata) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		organizationID, userID, action, resourceType, resourceID, requestID, ip, metadata,
	)
	return err
}
