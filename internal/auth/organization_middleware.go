package auth

import (
	"context"
	"net/http"
	"slices"

	"github.com/google/uuid"
)

type organizationContextKey struct{}

func WithOrganization(store *Store, roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok {
				http.Error(w, "authentication required", http.StatusUnauthorized)
				return
			}
			raw := r.Header.Get("X-Fuze-Organization")
			organizationID, err := uuid.Parse(raw)
			if err != nil {
				http.Error(w, "organization required", http.StatusBadRequest)
				return
			}
			org, err := store.OrganizationForUser(r.Context(), user.ID, organizationID)
			if err != nil {
				http.Error(w, "organization access denied", http.StatusForbidden)
				return
			}
			if len(roles) > 0 && !slices.Contains(roles, org.Role) {
				http.Error(w, "insufficient organization permissions", http.StatusForbidden)
				return
			}
			ctx := context.WithValue(r.Context(), organizationContextKey{}, org)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OrganizationFromContext(ctx context.Context) (Organization, bool) {
	org, ok := ctx.Value(organizationContextKey{}).(Organization)
	return org, ok
}
