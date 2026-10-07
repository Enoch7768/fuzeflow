package api

import (
	"net/http"

	"github.com/Enoch7768/fuzeflow/internal/auth"
	"github.com/Enoch7768/fuzeflow/internal/workflow"
)

type workflowValidationRequest struct {
	Workflow string `json:"workflow"`
}

func validateWorkflow(w http.ResponseWriter, r *http.Request) {
	var input workflowValidationRequest
	if err := decodeJSON(w, r, &input); err != nil {
		return
	}
	if len(input.Workflow) == 0 {
		http.Error(w, "workflow is required", http.StatusBadRequest)
		return
	}

	definition, err := workflow.Parse([]byte(input.Workflow))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := workflow.Validate(definition); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	plan, err := workflow.PlanExecution(definition)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"valid": true,
		"workflow": map[string]any{
			"name":    definition.Name,
			"version": definition.Version,
			"jobs":    len(definition.Jobs),
		},
		"plan": plan,
	})
}

func workflowOrganization(next http.Handler, store *auth.Store) http.Handler {
	return auth.WithSession(store, auth.WithOrganization(store)(auth.RequireAuth(next)))
}
