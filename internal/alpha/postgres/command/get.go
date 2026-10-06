package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/alpha/postgres"
	"github.com/nais/cli/internal/alpha/postgres/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

type postgresDetails struct {
	Name             string                   `json:"name"`
	Team             string                   `json:"team"`
	Environment      string                   `json:"environment"`
	Version          string                   `json:"version"`
	HighAvailability bool                     `json:"highAvailability"`
	Resources        postgresResourceRequests `json:"resources"`
	Labels           []postgresLabel          `json:"labels"`
	ActiveBranch     string                   `json:"activeBranch"`
	RequestedBranch  string                   `json:"requestedBranch,omitempty"`
	Branches         []postgresDetailsBranch  `json:"branches"`
}

type postgresResourceRequests struct {
	CPU      *string `json:"cpu"`
	Memory   *string `json:"memory"`
	DiskSize *string `json:"diskSize"`
}

type postgresLabel struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type postgresDetailsBranch struct {
	postgres.Branch
	Active    bool `json:"active"`
	Requested bool `json:"requested"`
}

func instanceGetCommand(parent *flag.Postgres) *naistrix.Command {
	f := &flag.BranchList{Postgres: parent}
	return &naistrix.Command{
		Name: "get", Title: "Get Postgres configuration and branches.",
		Description: "Show configured version, high availability, resource requests, labels, and branch observations. Omitted resource requests are not configured; defaults are not inferred. Use 'nais alpha postgres status' to check branch-based readiness.",
		Args:        []naistrix.Argument{{Name: "postgres"}}, Flags: f,
		AutoCompleteFunc: autoCompletePostgresNames(parent),
		Examples: []naistrix.Example{
			{Description: "Describe a Postgres in a specific environment.", Command: "my-postgres -t my-team -e dev-gcp"},
			{Description: "Resolve the environment of an existing Postgres and show configuration as JSON.", Command: "my-postgres -t my-team --output json"},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			env, err := branchEnvironment(ctx, parent, args.Get("postgres"))
			if err != nil {
				return err
			}
			p, err := postgres.GetPostgres(ctx, parent.Team, env, args.Get("postgres"))
			if err != nil {
				return err
			}
			detail := postgresDetails{
				Name: p.Name, Team: parent.Team, Environment: env, Version: p.MajorVersion, HighAvailability: p.HighAvailability,
				Resources: postgresResourceRequests{CPU: p.Resources.Cpu, Memory: p.Resources.Memory, DiskSize: p.Resources.DiskSize},
				Labels:    make([]postgresLabel, 0, len(p.Labels)),
				Branches:  make([]postgresDetailsBranch, 0, len(p.Branches.Nodes)),
			}
			status := postgres.BranchStatus{Branches: make([]postgres.Branch, 0, len(p.Branches.Nodes))}
			if p.ActiveBranch != nil {
				status.Active = p.ActiveBranch.Name
			}
			if p.DesiredActiveBranch != nil {
				status.DesiredActive = *p.DesiredActiveBranch
			}
			detail.ActiveBranch = status.Active
			if status.DesiredActive != "" && status.DesiredActive != status.Active {
				detail.RequestedBranch = status.DesiredActive
			}
			for _, label := range p.Labels {
				detail.Labels = append(detail.Labels, postgresLabel{Key: label.Key, Value: label.Value})
			}
			for _, branch := range p.Branches.Nodes {
				b := postgres.Branch{Name: branch.Name, State: branch.State}
				status.Branches = append(status.Branches, b)
				detail.Branches = append(detail.Branches, postgresDetailsBranch{
					Branch: b, Active: b.Name == status.Active,
					Requested: b.Name == detail.RequestedBranch,
				})
			}
			if f.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(detail)
			}
			requested := func(value *string) string {
				if value == nil {
					return "(not configured)"
				}
				return *value
			}
			active := detail.ActiveBranch
			if active == "" {
				active = "(none)"
			}
			rows := [][]string{
				{"Field", "Value"},
				{"Name", detail.Name},
				{"Team", detail.Team},
				{"Environment", detail.Environment},
				{"Version", detail.Version},
				{"High availability", fmt.Sprint(detail.HighAvailability)},
				{"Requested CPU", requested(detail.Resources.CPU)},
				{"Requested memory", requested(detail.Resources.Memory)},
				{"Requested disk size", requested(detail.Resources.DiskSize)},
				{"Active branch", active},
			}
			if detail.RequestedBranch != "" {
				rows = append(rows, []string{"Requested branch (activation pending)", detail.RequestedBranch})
			}
			out.Println("Postgres configuration")
			if err := out.Table(output.TableWithMargins()).Render(rows); err != nil {
				return err
			}
			out.Println("Labels")
			if len(detail.Labels) == 0 {
				out.Println("No labels configured.")
			} else if err := out.Table().Render(detail.Labels); err != nil {
				return err
			}
			out.Println("Branches")
			if len(status.Branches) == 0 {
				out.Println("No Postgres branches found.")
				return nil
			}
			return out.Table().Render(branchListRows(status))
		},
	}
}
