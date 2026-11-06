package postgres

import "github.com/nais/cli/internal/naisapi/gql"

// Branch retains the deprecated main-entry JSON shape for existing scripts.
// It is constructed from Postgres state, not read from a branch resource.
type Branch struct {
	Name  string            `json:"name"`
	State gql.PostgresState `json:"state"`
}

type BranchStatus struct {
	Active   string
	Branches []Branch
}
