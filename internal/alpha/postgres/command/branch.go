package command

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nais/cli/internal/alpha/postgres"
	"github.com/nais/cli/internal/alpha/postgres/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/input"
	"github.com/nais/naistrix/output"
)

func branchCommand(parent *flag.Postgres) *naistrix.Command {
	return &naistrix.Command{
		Name: "branch", Title: "Manage branches of a Nais Postgres (experimental).",
		SubCommands: []*naistrix.Command{
			branchListCommand(parent), branchStatusCommand(parent), branchCreateCommand(parent),
			branchActivateCommand(parent), branchDeleteCommand(parent),
		},
	}
}

func branchListCommand(parent *flag.Postgres) *naistrix.Command {
	f := &flag.BranchList{Postgres: parent}
	return &naistrix.Command{
		Name: "list", Title: "List branches of a Postgres.", Args: []naistrix.Argument{{Name: "postgres"}}, Flags: f,
		AutoCompleteFunc: autoCompletePostgresNames(parent),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			env, err := branchEnvironment(ctx, parent, args.Get("postgres"))
			if err != nil {
				return err
			}
			status, err := postgres.GetBranchStatus(ctx, parent.Team, env, args.Get("postgres"))
			if err != nil {
				return err
			}
			if f.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(status.Branches)
			}
			if len(status.Branches) == 0 {
				out.Println("No Postgres branches found.")
				return nil
			}
			return out.Table().Render(branchListRows(status))
		},
	}
}

type branchListRow struct {
	Name      string
	State     postgres.State
	Active    string
	Requested string
}

func branchListRows(status postgres.BranchStatus) []branchListRow {
	rows := make([]branchListRow, 0, len(status.Branches))
	for _, b := range status.Branches {
		row := branchListRow{Name: b.Name, State: postgres.State(b.State)}
		if b.Name == status.Active {
			row.Active = "Yes"
		}
		if b.Name == status.DesiredActive && status.DesiredActive != status.Active {
			row.Requested = "Yes"
		}
		rows = append(rows, row)
	}
	return rows
}

func branchStatusCommand(parent *flag.Postgres) *naistrix.Command {
	return &naistrix.Command{
		Name: "status", Title: "Show a branch's state and requested and observed activation.",
		Args: []naistrix.Argument{{Name: "postgres"}, {Name: "branch"}}, AutoCompleteFunc: autoCompletePostgresBranches(parent),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			env, err := branchEnvironment(ctx, parent, args.Get("postgres"))
			if err != nil {
				return err
			}
			status, err := postgres.GetNamedBranchStatus(ctx, parent.Team, env, args.Get("postgres"), args.Get("branch"))
			if err != nil {
				return err
			}
			printBranchStatus(out, status)
			return nil
		},
	}
}

func branchCreateCommand(parent *flag.Postgres) *naistrix.Command {
	f := &flag.BranchCreate{Postgres: parent}
	return &naistrix.Command{
		Name: "create", Title: "Create an inactive Postgres branch from a point in time.", Flags: f,
		Examples: []naistrix.Example{
			{Description: "Restore from the active branch as it was two hours ago.", Command: "my-postgres restored --ago 2h"},
			{Description: "Restore from the active branch at a UTC timestamp (Z means UTC).", Command: "my-postgres restored --at 2026-10-06T10:00:00Z"},
			{Description: "Restore from a specific source branch.", Command: "my-postgres restored --from main --at 2026-10-06T10:00:00Z"},
		},
		Args:             []naistrix.Argument{{Name: "postgres"}, {Name: "branch"}},
		AutoCompleteFunc: autoCompletePostgresNames(parent),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			at, err := resolveRecoveryTime(f.At, f.Ago, time.Now())
			if err != nil {
				return err
			}
			env, err := branchEnvironment(ctx, parent, args.Get("postgres"))
			if err != nil {
				return err
			}
			source := string(f.From)
			if source == "" {
				status, err := postgres.GetBranchStatus(ctx, parent.Team, env, args.Get("postgres"))
				if err != nil {
					return err
				}
				source = status.Active
				if source == "" {
					return fmt.Errorf("no active branch is known for Postgres %q; specify a source branch with --from", args.Get("postgres"))
				}
			}
			b, err := postgres.CreateBranch(ctx, gql.CreatePostgresBranchInput{
				Postgres: args.Get("postgres"), Branch: args.Get("branch"), SourceBranch: source,
				TargetTime: at, EnvironmentName: env, TeamSlug: parent.Team,
			})
			if err != nil {
				return err
			}
			out.Printf("Branch %q created; observed state: %s. Provisioning may still be in progress.\n", b.Name, postgres.State(b.State).String())
			return nil
		},
	}
}

func branchActivateCommand(parent *flag.Postgres) *naistrix.Command {
	f := &flag.BranchActivate{Postgres: parent}
	return &naistrix.Command{
		Name: "activate", Title: "Request activation of a Postgres branch.", Flags: f,
		Args:             []naistrix.Argument{{Name: "postgres"}, {Name: "branch"}},
		AutoCompleteFunc: autoCompletePostgresBranches(parent),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			env, err := branchEnvironment(ctx, parent, args.Get("postgres"))
			if err != nil {
				return err
			}
			if !f.Yes {
				ok, err := input.Confirm(fmt.Sprintf("Activate branch %q of Postgres %q in %q?", args.Get("branch"), args.Get("postgres"), env))
				if err != nil {
					return err
				}
				if !ok {
					return fmt.Errorf("cancelled by user")
				}
			}
			status, err := postgres.ActivateBranch(ctx, gql.ActivatePostgresBranchInput{
				Postgres: args.Get("postgres"), Branch: args.Get("branch"), EnvironmentName: env, TeamSlug: parent.Team,
			})
			if err != nil {
				return err
			}
			out.Println("Activation requested; reconciliation may still be in progress.")
			printBranchStatus(out, status)
			return nil
		},
	}
}

func branchDeleteCommand(parent *flag.Postgres) *naistrix.Command {
	f := &flag.BranchDelete{Postgres: parent}
	return &naistrix.Command{
		Name: "delete", Title: "Delete an inactive Postgres branch.", Flags: f,
		Args:             []naistrix.Argument{{Name: "postgres"}, {Name: "branch"}},
		AutoCompleteFunc: autoCompletePostgresBranches(parent),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			env, err := branchEnvironment(ctx, parent, args.Get("postgres"))
			if err != nil {
				return err
			}
			if !f.Yes {
				ok, err := input.Confirm(fmt.Sprintf("Delete branch %q of Postgres %q in %q?", args.Get("branch"), args.Get("postgres"), env))
				if err != nil {
					return err
				}
				if !ok {
					return fmt.Errorf("cancelled by user")
				}
			}
			deleted, err := postgres.DeleteBranch(ctx, gql.DeletePostgresBranchInput{
				Postgres: args.Get("postgres"), Branch: args.Get("branch"), EnvironmentName: env, TeamSlug: parent.Team,
			})
			if err != nil {
				return err
			}
			if !deleted {
				return fmt.Errorf("branch %q was not deleted", args.Get("branch"))
			}
			out.Printf("Deleted branch %q of Postgres %q in %q.\n", args.Get("branch"), args.Get("postgres"), env)
			return nil
		},
	}
}

func branchEnvironment(ctx context.Context, f *flag.Postgres, name string) (string, error) {
	if f.Team == "" {
		return "", fmt.Errorf("missing required team, specify -t, --team or set a default team")
	}
	return resolvePostgresEnvironment(ctx, f.Team, name, string(f.Environment))
}

func printBranchStatus(out *naistrix.OutputWriter, status postgres.BranchStatus) {
	value := func(s string) string {
		if s == "" {
			return "(none)"
		}
		return s
	}
	out.Printf("Requested active branch: %s\nObserved active branch: %s\n", value(status.DesiredActive), value(status.Active))
	if status.DesiredActive != "" && status.DesiredActive != status.Active {
		out.Println("Activation pending: requested and observed branches differ.")
	}
	for _, b := range status.Branches {
		out.Printf("%s\t%s\n", b.Name, postgres.State(b.State).String())
	}
}

func resolveRecoveryTime(at, ago string, now time.Time) (time.Time, error) {
	if at != "" && ago != "" {
		return time.Time{}, fmt.Errorf("--at and --ago cannot be used together")
	}
	if at != "" {
		return parseRecoveryTime(at)
	}
	if ago == "" {
		return time.Time{}, fmt.Errorf("--at is required unless --ago is specified (e.g. --at 2026-10-06T10:00:00Z or --ago 2h)")
	}
	duration, err := time.ParseDuration(ago)
	if err != nil || duration <= 0 {
		return time.Time{}, fmt.Errorf("--ago must be a positive duration, e.g. 2h, 30m or 1h30m")
	}
	return now.Add(-duration).UTC().Truncate(time.Second), nil
}

func parseRecoveryTime(value string) (time.Time, error) {
	at, err := time.Parse(time.RFC3339, value)
	if err != nil || !strings.HasSuffix(value, "Z") {
		return time.Time{}, fmt.Errorf("--at must be an RFC3339 UTC timestamp ending in Z")
	}
	if at.After(time.Now()) {
		return time.Time{}, fmt.Errorf("--at must not be in the future")
	}
	if at.Nanosecond() != 0 {
		return time.Time{}, fmt.Errorf("--at must have whole-second precision")
	}
	return at, nil
}
