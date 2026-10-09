package command

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nais/cli/internal/alpha/postgres"
	"github.com/nais/cli/internal/alpha/postgres/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/input"
	"github.com/nais/naistrix/output"
)

func branchCommand(parent *flag.Postgres) *naistrix.Command {
	return &naistrix.Command{
		Name: "branch", Title: "Manage branches of a Nais Postgres (experimental).",
		Description: "Create, list, inspect, activate and delete branches of a Postgres. A branch is a copy of the database restored from a point in time. Only one branch is active at a time; workloads can also select a branch explicitly with uses.postgres[].branch.",
		SubCommands: []*naistrix.Command{
			branchListCommand(parent), branchCreateCommand(parent),
			branchActivateCommand(parent), branchDeleteCommand(parent),
		},
	}
}

func branchListCommand(parent *flag.Postgres) *naistrix.Command {
	f := &flag.BranchList{Postgres: parent}
	return &naistrix.Command{
		Name: "list", Title: "List branches of a Postgres.",
		Description: "List the branches of a Postgres with their status and whether each one is active.",
		Args:        []naistrix.Argument{{Name: "postgres", Prompt: "Name of the Postgres instance to list branches for"}}, Flags: f,
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
			printPostgresHeading(out, args.Get("postgres"), env)
			printPendingActivation(out, status)
			if len(status.Branches) == 0 {
				out.Println("No Postgres branches found.")
				return nil
			}
			return out.Table().Render(branchListRows(status))
		},
	}
}

type branchListRow struct {
	Name   string
	Status postgres.State
	Active string
}

func branchListRows(status postgres.BranchStatus) []branchListRow {
	rows := make([]branchListRow, 0, len(status.Branches))
	for _, b := range status.Branches {
		row := branchListRow{Name: b.Name, Status: postgres.State(b.State), Active: "No"}
		if b.Name == status.Active {
			row.Active = "Yes"
		}
		rows = append(rows, row)
	}
	return rows
}

func branchCreateCommand(parent *flag.Postgres) *naistrix.Command {
	f := &flag.BranchCreate{Postgres: parent}
	return &naistrix.Command{
		Name: "create", Title: "Create an inactive Postgres branch from a point in time.", Flags: f,
		Description: "Use create <postgres> [<new-branch>]. If the new branch name is omitted, prompt for it interactively. An explicit or default environment is used directly; otherwise resolve from the named Postgres, auto-selecting its sole environment or prompting for multiple (require -e noninteractively when ambiguous).",
		Examples: []naistrix.Example{
			{Description: "Prompt for a new branch name and restore from five minutes ago.", Command: "my-postgres --ago 5m"},
			{Description: "Restore from the active branch as it was two hours ago.", Command: "my-postgres restored --ago 2h"},
			{Description: "Restore from the active branch at a UTC timestamp (Z means UTC).", Command: "my-postgres restored --at 2026-10-06T10:00:00Z"},
			{Description: "Restore from a specific source branch.", Command: "my-postgres restored --from main --at 2026-10-06T10:00:00Z"},
		},
		Args:             []naistrix.Argument{{Name: "postgres", Prompt: "Name of the Postgres instance; optionally add a new branch name next", Repeatable: true}},
		AutoCompleteFunc: autoCompletePostgresNames(parent),
		ValidateFunc: naistrix.ValidateFuncs(
			validation.RequireTeam(f),
			func(_ context.Context, args *naistrix.Arguments) error {
				values := args.All()
				if len(values) < 1 || len(values) > 2 {
					return fmt.Errorf("expected a Postgres and optional new branch: create <postgres> [<new-branch>]")
				}
				if strings.TrimSpace(values[0]) == "" {
					return fmt.Errorf("Postgres name must not be empty")
				}
				if len(values) == 2 && strings.TrimSpace(values[1]) == "" {
					return fmt.Errorf("new branch name must not be empty")
				}
				return nil
			},
		),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			at, err := resolveRecoveryTime(f.At, f.Ago, time.Now())
			if err != nil {
				return err
			}
			values := args.All()
			name := values[0]
			var branch string
			if len(values) == 2 {
				branch = values[1]
			} else {
				entered, err := input.Input("Name of the new branch")
				if errors.Is(err, input.ErrNotInteractive) {
					return fmt.Errorf("specify the new branch name as the second positional argument, e.g. nais alpha postgres branch create mypg 5m-restore --ago 5m")
				}
				if err != nil {
					return fmt.Errorf("prompting for new branch name: %w", err)
				}
				branch = strings.TrimSpace(entered)
				if branch == "" {
					return fmt.Errorf("new branch name must not be empty")
				}
			}
			env, err := branchEnvironment(ctx, parent, name)
			if err != nil {
				return err
			}
			source := string(f.From)
			if source == "" {
				status, err := postgres.GetBranchStatus(ctx, parent.Team, env, name)
				if err != nil {
					return err
				}
				source = status.Active
				if source == "" {
					return fmt.Errorf("no active branch is known for Postgres %q; specify a source branch with --from", name)
				}
			}
			b, err := postgres.CreateBranch(ctx, gql.CreatePostgresBranchInput{
				Postgres: name, Branch: branch, SourceBranch: source,
				TargetTime: at, EnvironmentName: env, TeamSlug: parent.Team,
			})
			if err != nil {
				return err
			}
			out.Printf("Started creating branch %q. Run this command to check status:\n", b.Name)
			out.Println(branchListCommandLine(name, parent.Team, env, parent.Config))
			return nil
		},
	}
}

func branchActivateCommand(parent *flag.Postgres) *naistrix.Command {
	f := &flag.BranchActivate{Postgres: parent}
	return &naistrix.Command{
		Name: "activate", Title: "Request activation of a Postgres branch.", Flags: f,
		Description:      "Request that a branch becomes the active branch of the Postgres. Activation continues asynchronously; use branch list to check progress.",
		Args:             []naistrix.Argument{{Name: "postgres", Prompt: "Name of the Postgres instance"}, {Name: "branch", Prompt: "Name of the branch to activate"}},
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
			printPostgresHeading(out, args.Get("postgres"), env)
			if status.DesiredActive != "" && status.DesiredActive != status.Active {
				printPendingActivation(out, status)
			} else if status.Active == args.Get("branch") {
				out.Printf("Branch %q is active.\n", status.Active)
			} else {
				out.Printf("Started activating branch %q.\n", args.Get("branch"))
			}
			out.Println("Run this command to check status:")
			out.Println(branchListCommandLine(args.Get("postgres"), parent.Team, env, parent.Config))
			return nil
		},
	}
}

func branchDeleteCommand(parent *flag.Postgres) *naistrix.Command {
	f := &flag.BranchDelete{Postgres: parent}
	return &naistrix.Command{
		Name: "delete", Title: "Delete an inactive Postgres branch.", Flags: f,
		Description:      "Delete an inactive branch, even if workloads still use it. Referring apps/jobs are listed with a warning before confirmation. --yes skips confirmation, not the usage check or warning.",
		Args:             []naistrix.Argument{{Name: "postgres", Prompt: "Name of the Postgres instance"}, {Name: "branch", Prompt: "Name of the inactive branch to delete"}},
		AutoCompleteFunc: autoCompletePostgresBranches(parent),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			env, err := branchEnvironment(ctx, parent, args.Get("postgres"))
			if err != nil {
				return err
			}
			workloads, err := postgres.GetBranchWorkloads(ctx, parent.Team, env, args.Get("postgres"), args.Get("branch"))
			if err != nil {
				return err
			}
			if len(workloads) > 0 {
				out.Warnf("Branch %q of Postgres %q in %q is used by the following workloads; deleting it will make their database connections unavailable:\n", args.Get("branch"), args.Get("postgres"), env)
				if err := out.Table().Render(workloads); err != nil {
					return err
				}
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
			out.Printf("Started deleting branch %q of Postgres %q in %q.\n", args.Get("branch"), args.Get("postgres"), env)
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

func printPendingActivation(out *naistrix.OutputWriter, status postgres.BranchStatus) {
	if status.DesiredActive == "" || status.DesiredActive == status.Active {
		return
	}
	current := status.Active
	if current == "" {
		current = "unknown"
	}
	out.Printf("Waiting to activate %s. Current active branch: %s.\n", status.DesiredActive, current)
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
