package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/internal/alpha/postgres"
	"github.com/nais/cli/internal/alpha/postgres/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

const postgresStatusScope = "Readiness is derived from branch states only; it does not verify workload switchover, SQL connectivity, or completion of resource-setting updates."

type postgresStatusResult struct {
	Name            string                 `json:"name"`
	Team            string                 `json:"team"`
	Environment     string                 `json:"environment"`
	Status          string                 `json:"status"`
	Reason          string                 `json:"reason,omitempty"`
	ActiveBranch    string                 `json:"activeBranch"`
	RequestedBranch string                 `json:"requestedBranch,omitempty"`
	Branches        []postgresStatusBranch `json:"branches"`
	Scope           string                 `json:"scope"`
}

type postgresStatusBranch struct {
	postgres.Branch
	Active    bool `json:"active"`
	Requested bool `json:"requested"`
}

func instanceStatusCommand(parent *flag.Postgres) *naistrix.Command {
	f := &flag.BranchList{Postgres: parent}
	return &naistrix.Command{
		Name: "status", Title: "Show Postgres readiness and branch activation.",
		Description: postgresStatusScope,
		Args:        []naistrix.Argument{{Name: "postgres"}}, Flags: f,
		AutoCompleteFunc: autoCompletePostgresNames(parent),
		Examples: []naistrix.Example{
			{Description: "Check provisioning progress or readiness.", Command: "my-postgres -t my-team -e dev-gcp"},
			{Description: "Resolve the environment of an existing Postgres and show status as JSON.", Command: "my-postgres -t my-team --output json"},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			name := args.Get("postgres")
			env, err := branchEnvironment(ctx, parent, name)
			if err != nil {
				return err
			}
			status, err := postgres.GetBranchStatus(ctx, parent.Team, env, name)
			if err != nil {
				return err
			}
			state, reason := summarizePostgresStatus(status)
			rows := branchListRows(status)
			result := postgresStatusResult{
				Name: name, Team: parent.Team, Environment: env, Status: state, Reason: reason,
				ActiveBranch: status.Active, Scope: postgresStatusScope,
				Branches: make([]postgresStatusBranch, 0, len(rows)),
			}
			if status.DesiredActive != "" && status.DesiredActive != status.Active {
				result.RequestedBranch = status.DesiredActive
			}
			for _, row := range rows {
				result.Branches = append(result.Branches, postgresStatusBranch{
					Branch: postgres.Branch{Name: row.Name, State: gql.PostgresBranchState(row.State)},
					Active: row.Active == "Yes", Requested: row.Requested == "Yes",
				})
			}
			if f.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(result)
			}
			out.Printf("Postgres %q in %q\nStatus: %s\n", name, env, state)
			if reason != "" {
				out.Printf("Reason: %s\n", reason)
			}
			active := status.Active
			if active == "" {
				active = "(none)"
			}
			out.Printf("Active branch: %s\n", active)
			if result.RequestedBranch != "" {
				out.Printf("Requested branch: %s (activation pending)\n", result.RequestedBranch)
			}
			out.Println(postgresStatusScope)
			if len(rows) == 0 {
				out.Println("No Postgres branches found.")
				return nil
			}
			return out.Table().Render(rows)
		},
	}
}

func summarizePostgresStatus(status postgres.BranchStatus) (string, string) {
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
			return "Degraded", fmt.Sprintf("Active branch %q is degraded.", status.Active)
		case gql.PostgresBranchStateAvailable, gql.PostgresBranchStateProgressing:
		default:
			return "Unknown", fmt.Sprintf("Active branch %q has no known state.", status.Active)
		}
	}
	if status.DesiredActive != "" && status.DesiredActive != status.Active {
		if desiredState == gql.PostgresBranchStateDegraded {
			return "Degraded", fmt.Sprintf("Requested branch %q is degraded; activation is pending.", status.DesiredActive)
		}
		return "Progressing", fmt.Sprintf("Activation of requested branch %q is pending.", status.DesiredActive)
	}
	switch activeState {
	case gql.PostgresBranchStateAvailable:
		return "Ready", ""
	case gql.PostgresBranchStateProgressing:
		return "Progressing", fmt.Sprintf("Active branch %q is progressing.", status.Active)
	default:
		return "Unknown", "No observed active branch is known."
	}
}

func quoteShellArgument(value string) string {
	if value != "" && strings.Trim(value, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_.") == "" {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func postgresStatusCommandLine(name, team, environment string) string {
	return fmt.Sprintf("nais alpha postgres status %s -t %s -e %s", quoteShellArgument(name), quoteShellArgument(team), quoteShellArgument(environment))
}
