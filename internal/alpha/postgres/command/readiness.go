package command

import (
	"github.com/nais/cli/internal/alpha/postgres"
	"github.com/nais/cli/internal/naisapi/gql"
)

func postgresReadiness(status postgres.BranchStatus) string {
	var activeState, desiredState gql.PostgresBranchState
	for _, branch := range status.Branches {
		if branch.Name == status.Active {
			activeState = branch.State
		}
		if branch.Name == status.DesiredActive {
			desiredState = branch.State
		}
	}
	if status.Active != "" {
		switch activeState {
		case gql.PostgresBranchStateDegraded:
			return "Needs attention"
		case gql.PostgresBranchStateAvailable, gql.PostgresBranchStateProgressing:
		default:
			return "Unknown"
		}
	}
	if status.DesiredActive != "" && status.DesiredActive != status.Active {
		if desiredState == gql.PostgresBranchStateDegraded {
			return "Needs attention"
		}
		if status.Active == "" {
			return "Not ready yet"
		}
		return "Updating"
	}
	switch activeState {
	case gql.PostgresBranchStateAvailable:
		return "Ready"
	case gql.PostgresBranchStateProgressing:
		return "Not ready yet"
	default:
		return "Unknown"
	}
}
