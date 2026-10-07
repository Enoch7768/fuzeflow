package api

import (
	"net/http"

	"github.com/Enoch7768/fuzeflow/internal/auth"
)

func (a *authAPI) organizations(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	organizations, err := a.store.OrganizationsForUser(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "unable to load organizations", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, organizations)
}
