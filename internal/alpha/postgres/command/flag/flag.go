package flag

import (
	"context"
	"sort"
	"strings"
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
		Reason      string        `name:"reason" usage:"Reason for personal access (at least 10 characters)."`
		TTL         time.Duration `name:"ttl" usage:"Requested lifetime (default 30m, maximum 1h)."`
		Database    string        `name:"database" usage:"Database name for psql (default app)."`
	}
	Proxy struct {
		*Postgres
		Branch        string        `name:"branch" usage:"Branch to access (defaults to the active branch)."`
		AccessLevel   string        `name:"access-level" usage:"Access level: read, write, or admin."`
		Reason        string        `name:"reason" usage:"Reason for personal access (at least 10 characters)."`
		TTL           time.Duration `name:"ttl" usage:"Requested lifetime (default 30m, maximum 1h)."`
		Host          string        `name:"host" usage:"Local loopback address (default 127.0.0.1)."`
		Port          int           `name:"port" usage:"Local port (default random)."`
		PrintPassword bool          `name:"print-password" usage:"Print the database password to stdout (sensitive)."`
	}
	BranchList struct {
		*Postgres
		Output Output `name:"output" short:"o" usage:"Format output (table or json)."`
	}
	BranchCreate struct {
		*Postgres
		From BranchSource `name:"from" usage:"Source branch to recover from (required)."`
		At   string       `name:"at" usage:"UTC recovery point in RFC3339 format ending in Z (required)."`
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

func (*List) LabelFacetResource() string { return "postgresBranches" }

// BranchSource completes source branches for the selected Postgres and environment.
type BranchSource string

var _ naistrix.FlagAutoCompleter = (*BranchSource)(nil)

func (s *BranchSource) AutoComplete(ctx context.Context, args *naistrix.Arguments, _ string, flags any) ([]string, string) {
	f := flags.(*BranchCreate)
	if f.Team == "" || args.Get("postgres") == "" {
		return nil, "Select a Postgres first (and specify a team)."
	}
	instances, err := postgres.GetTeamPostgresBranches(ctx, f.Team, nil, nil)
	if err != nil {
		return nil, "Unable to fetch Postgres branches."
	}
	name := args.Get("postgres")
	environment := string(f.Environment)
	if environment == "" {
		envs := make(map[string]bool)
		for _, instance := range instances {
			pg, _, _ := strings.Cut(instance.Name.Name, "/")
			if pg == name {
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
	var branches []string
	for _, instance := range instances {
		pg, branch, ok := strings.Cut(instance.Name.Name, "/")
		if ok && pg == name && instance.Environment == environment {
			branches = append(branches, branch)
		}
	}
	sort.Strings(branches)
	return branches, ""
}

type Output string

var _ naistrix.FlagAutoCompleter = (*Output)(nil)

func (o *Output) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"table", "json"}, "Available output formats."
}
