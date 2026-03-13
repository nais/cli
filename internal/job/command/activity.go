package command

import (
	"context"
	"os"
	"strings"

	"github.com/nais/cli/internal/job"
	"github.com/nais/cli/internal/job/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func activity(parentFlags *flag.Job) *naistrix.Command {
	flags := &flag.Activity{
		Job:    parentFlags,
		Output: "table",
		Limit:  20,
	}

	return &naistrix.Command{
		Name:  "activity",
		Title: "Show activity for a job.",
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			ret, err := job.GetJobActivity(ctx, flags.Team, args.Get("name"), flags.Environment, flags.Limit)
			if err != nil {
				return err
			}
			if len(ret) == 0 {
				out.Println("No activity found for job.")
				return nil
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(ret)
			}

			return out.Table().Render(ret)
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() == 0 {
				if len(flags.Team) == 0 {
					return nil, "Please provide team to auto-complete job names. 'nais config set team <team>', or '--team <team>' flag."
				}
				environments := flags.Environment
				if len(environments) == 0 {
					environments = environmentsFromCLIArgs()
				}
				if len(environments) == 0 {
					return nil, "Please provide environment to auto-complete job names. '--environment <environment>' flag."
				}

				jobs, err := job.GetJobNames(ctx, flags.Team, environments)
				if err != nil {
					return nil, "Unable to fetch job names."
				}
				return jobs, "Select a job."
			}
			return nil, ""
		},
	}
}

func environmentsFromCLIArgs() []string {
	seen := map[string]struct{}{}
	environments := make([]string, 0)
	args := os.Args

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-e" || arg == "--environment":
			if i+1 >= len(args) {
				continue
			}
			next := args[i+1]
			if strings.HasPrefix(next, "-") || next == "" {
				continue
			}
			if _, ok := seen[next]; !ok {
				seen[next] = struct{}{}
				environments = append(environments, next)
			}
			i++
		case strings.HasPrefix(arg, "--environment="):
			env := strings.TrimPrefix(arg, "--environment=")
			if env != "" {
				if _, ok := seen[env]; !ok {
					seen[env] = struct{}{}
					environments = append(environments, env)
				}
			}
		case strings.HasPrefix(arg, "-e="):
			env := strings.TrimPrefix(arg, "-e=")
			if env != "" {
				if _, ok := seen[env]; !ok {
					seen[env] = struct{}{}
					environments = append(environments, env)
				}
			}
		}
	}

	return environments
}
