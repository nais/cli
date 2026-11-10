package postgres

import "github.com/nais/cli/internal/naisapi/gql"

// Branch describes the observed state of a named branch.
type Branch struct {
	Name  string                  `json:"name"`
	State gql.PostgresBranchState `json:"state"`
}

// BranchStatus distinguishes requested activation from the observed active branch.
// Branch observations are used internally to summarize instance readiness.
type BranchStatus struct {
	DesiredActive string
	Active        string
	Branches      []Branch
}
