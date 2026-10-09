package command

import (
	"context"
	"strings"

	"github.com/nais/cli/internal/alpha/postgres"
	"github.com/nais/cli/internal/alpha/postgres/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
	"github.com/pterm/pterm"
)

type postgresDetails struct {
	Name             string                   `json:"name"`
	Team             string                   `json:"team"`
	Environment      string                   `json:"environment"`
	Version          string                   `json:"version"`
	Status           string                   `json:"status"`
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

func printPostgresHeading(out *naistrix.OutputWriter, name, environment string) {
	out.Printf("%s · %s\n", pterm.NewStyle(pterm.Bold).Sprint(name), pterm.FgGray.Sprint(environment))
}

func instanceGetCommand(parent *flag.Postgres) *naistrix.Command {
	f := &flag.BranchList{Postgres: parent}
	return &naistrix.Command{
		Name: "get", Title: "Get Postgres details and readiness.",
		Description: "Show configured settings and a coarse platform-reported status, not SQL connectivity, workload readiness, or completion of end-to-end resource updates. Omitted resource requests are not configured; defaults are not inferred. Detailed provisioning stages are unavailable. Use branch commands to inspect branches.",
		Args:        []naistrix.Argument{{Name: "postgres", Prompt: "Name of the Postgres instance to inspect"}}, Flags: f,
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
			detail.Status = postgresReadiness(status)
			if f.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(detail)
			}
			requested := func(value *string) string {
				if value == nil {
					return "(not configured)"
				}
				return *value
			}
			printPostgresHeading(out, detail.Name, detail.Environment)
			style := pterm.NewStyle()
			switch detail.Status {
			case "Ready":
				style = pterm.NewStyle(pterm.FgGreen)
			case "Not ready yet", "Updating":
				style = pterm.NewStyle(pterm.FgYellow)
			case "Needs attention":
				style = pterm.NewStyle(pterm.FgRed)
			}
			out.Printf("  %-18s %s\n\n", "Status:", style.Sprint(detail.Status))
			ha := "No"
			if detail.HighAvailability {
				ha = "Yes"
			}
			for _, setting := range []struct{ label, value string }{
				{"Version:", detail.Version},
				{"High availability:", ha},
				{"CPU:", requested(detail.Resources.CPU)},
				{"Memory:", requested(detail.Resources.Memory)},
				{"Disk:", requested(detail.Resources.DiskSize)},
			} {
				out.Printf("  %-18s %s\n", setting.label, setting.value)
			}
			if len(detail.Labels) > 0 {
				labels := make([]string, 0, len(detail.Labels))
				for _, label := range detail.Labels {
					labels = append(labels, label.Key+"="+label.Value)
				}
				out.Printf("\n  %-18s %s\n", "Labels:", strings.Join(labels, ", "))
			}
			return nil
		},
	}
}
