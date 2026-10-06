package flag

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/nais/cli/internal/alpha/postgres"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/labels"
	"github.com/nais/naistrix"
)

type (
	Postgres struct{ *flags.GlobalFlags }
	Access   struct {
		*Postgres
		Branch      string        `name:"branch" usage:"Branch to access (defaults to the active branch)."`
		AccessLevel string        `name:"access-level" usage:"Access level: read, write, or admin."`
		Reason      string        `name:"reason" usage:"Reason for personal access (at least 10 characters); prompted if omitted interactively."`
		TTL         time.Duration `name:"ttl" usage:"Requested lifetime (default 30m, maximum 1h)."`
		Database    string        `name:"database" usage:"Database name for psql (default app)."`
	}
	Proxy struct {
		*Postgres
		Branch        string        `name:"branch" usage:"Branch to access (defaults to the active branch)."`
		AccessLevel   string        `name:"access-level" usage:"Access level: read, write, or admin."`
		Reason        string        `name:"reason" usage:"Reason for personal access (at least 10 characters); prompted if omitted interactively."`
		TTL           time.Duration `name:"ttl" usage:"Requested lifetime (default 30m, maximum 1h)."`
		Host          string        `name:"host" usage:"Local loopback address (default 127.0.0.1)."`
		Port          int           `name:"port" usage:"Local port (default random)."`
		PrintPassword bool          `name:"print-password" usage:"Print the database password to stdout (sensitive)."`
	}
	InstanceCreate struct {
		*Postgres
		Yes              bool            `name:"yes" short:"y" usage:"Confirm without a prompt."`
		Version          PostgresVersion `name:"version" usage:"PostgreSQL major version (18)."`
		HighAvailability bool            `name:"high-availability" usage:"Enable high availability (third instance and synchronous replication)."`
		CPU              string          `name:"cpu" usage:"Requested CPU, e.g. 100m."`
		Memory           string          `name:"memory" usage:"Requested memory, e.g. 512Mi."`
		DiskSize         string          `name:"disk-size" usage:"Requested disk size, e.g. 10Gi."`
	}
	InstanceDelete struct {
		*Postgres
		Yes bool `name:"yes" short:"y" usage:"Confirm irreversible deletion without a prompt."`
	}
	InstanceUpdate struct {
		*Postgres
		Yes              bool             `name:"yes" short:"y" usage:"Confirm without a prompt."`
		HighAvailability HighAvailability `name:"high-availability" usage:"Enable or disable high availability (true or false)."`
		CPU              string           `name:"cpu" usage:"Requested CPU, e.g. 100m."`
		Memory           string           `name:"memory" usage:"Requested memory, e.g. 512Mi."`
		DiskSize         string           `name:"disk-size" usage:"Requested disk size, e.g. 10Gi."`
	}
	BranchList struct {
		*Postgres
		Output Output `name:"output" short:"o" usage:"Format output (table or json)."`
	}
	BranchCreate struct {
		*Postgres
		From BranchSource `name:"from" usage:"Source branch to recover from (defaults to the active branch)."`
		At   string       `name:"at" usage:"UTC recovery point, e.g. 2026-10-06T10:00:00Z (use either --at or --ago)."`
		Ago  string       `name:"ago" usage:"Recover from this long ago, e.g. 2h, 30m or 1h30m (use either --at or --ago)."`
	}
	BranchActivate struct {
		*Postgres
		Yes bool `name:"yes" short:"y" usage:"Confirm activation without a prompt."`
	}
	BranchDelete struct {
		*Postgres
		Yes bool `name:"yes" short:"y" usage:"Confirm deletion without a prompt."`
	}
	List struct {
		*Postgres
		Output Output              `name:"output" short:"o" usage:"Format output (table or json)."`
		Labels labels.LabelFilters `name:"label" short:"l" usage:"Filter by label in |KEY=VALUE| form. Can be repeated."`
	}
)

func (*List) LabelFacetResource() string { return "postgreses" }

// BranchSource completes source branches for the selected Postgres and environment.
type BranchSource string

var _ naistrix.FlagAutoCompleter = (*BranchSource)(nil)

func (s *BranchSource) AutoComplete(ctx context.Context, args *naistrix.Arguments, _ string, flags any) ([]string, string) {
	f := flags.(*BranchCreate)
	values := args.All()
	if f.Team == "" || len(values) == 0 || values[0] == "" {
		return nil, "Select a Postgres first (and specify a team)."
	}
	name := values[0]
	environment := string(f.Environment)
	if environment == "" {
		instances, err := postgres.GetTeamPostgreses(ctx, f.Team, nil, nil)
		if err != nil {
			return nil, fmt.Sprintf("Unable to fetch Postgres environments: %v", err)
		}
		envs := make(map[string]bool)
		for _, instance := range instances {
			if instance.Name.Name == name {
				envs[instance.Environment] = true
			}
		}
		if len(envs) != 1 {
			return nil, "Specify --environment to complete source branches."
		}
		for env := range envs {
			environment = env
		}
	}
	status, err := postgres.GetBranchStatus(ctx, f.Team, environment, name)
	if err != nil {
		return nil, fmt.Sprintf("Unable to fetch Postgres branches: %v", err)
	}
	branches := make([]string, 0, len(status.Branches))
	for _, branch := range status.Branches {
		branches = append(branches, branch.Name)
	}
	sort.Strings(branches)
	return branches, ""
}

// PostgresVersion completes supported major versions for creation.
type PostgresVersion string

var _ naistrix.FlagAutoCompleter = (*PostgresVersion)(nil)

func (*PostgresVersion) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"18"}, "Supported PostgreSQL versions."
}

// HighAvailability completes explicit update values, including false.
type HighAvailability string

var _ naistrix.FlagAutoCompleter = (*HighAvailability)(nil)

func (*HighAvailability) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"true", "false"}, "High availability values."
}

type Output string

var _ naistrix.FlagAutoCompleter = (*Output)(nil)

func (o *Output) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"table", "json"}, "Available output formats."
}
