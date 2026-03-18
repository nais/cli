package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/config"
	"github.com/nais/cli/internal/config/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
	"github.com/pterm/pterm"
)

// Entry represents a key-value pair in a config.
type Entry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ConfigDetail struct {
	Name         string              `json:"name"`
	Environment  string              `json:"environment"`
	Data         []Entry             `json:"data"`
	LastModified config.LastModified `json:"lastModified"`
	ModifiedBy   string              `json:"modifiedBy,omitempty"`
	Workloads    []string            `json:"workloads,omitempty"`
}

func get(parentFlags *flag.Config) *naistrix.Command {
	f := &flag.Get{Config: parentFlags}
	return &naistrix.Command{
		Name:        "get",
		Title:       "Get details about a config.",
		Description: "This command shows details about a config, including its key-value pairs, workloads using it, and last modification info.",
		Flags:       f,
		Args:        defaultArgs,
		ValidateFunc: func(_ context.Context, args *naistrix.Arguments) error {
			if err := validateSingleEnvironmentFlagUsage(); err != nil {
				return err
			}
			if providedEnvironment := string(f.Environment); providedEnvironment != "" {
				if err := validation.CheckEnvironment(providedEnvironment); err != nil {
					return err
				}
			}
			return validateArgs(args)
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() == 0 {
				return autoCompleteConfigNames(ctx, f.Team, string(f.Environment), false)
			}
			return nil, ""
		},
		Examples: []naistrix.Example{
			{
				Description: "Get details for a config named my-config in environment dev.",
				Command:     "my-config --environment dev",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			if providedEnvironment := string(f.Environment); providedEnvironment != "" {
				return runGetCommand(ctx, args, out, f.Team, providedEnvironment, f.Output)
			}

			environment, err := resolveConfigEnvironment(ctx, f.Team, args.Get("name"), string(f.Environment))
			if err != nil {
				return err
			}

			return runGetCommand(ctx, args, out, f.Team, environment, f.Output)
		},
	}
}

func runGetCommand(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter, team, environment string, outputFormat flag.Output) error {
	metadata := metadataFromArgs(args, team, environment)

	existing, err := config.Get(ctx, metadata)
	if err != nil {
		return fmt.Errorf("fetching config: %w", err)
	}

	entries := make([]Entry, len(existing.Values))
	for i, v := range existing.Values {
		entries[i] = Entry{Key: v.Name, Value: v.Value}
	}

	if outputFormat == "json" {
		detail := ConfigDetail{
			Name:         existing.Name,
			Environment:  existing.TeamEnvironment.Environment.Name,
			Data:         entries,
			LastModified: config.LastModified(existing.LastModifiedAt),
		}
		if existing.LastModifiedBy.Email != "" {
			detail.ModifiedBy = existing.LastModifiedBy.Email
		}
		for _, w := range existing.Workloads.Nodes {
			detail.Workloads = append(detail.Workloads, w.GetName())
		}
		return out.JSON(output.JSONWithPrettyOutput()).Render(detail)
	}

	pterm.DefaultSection.Println("Config details")
	err = pterm.DefaultTable.
		WithHasHeader().
		WithHeaderRowSeparator("-").
		WithData(config.FormatDetails(metadata, existing)).
		Render()
	if err != nil {
		return fmt.Errorf("rendering table: %w", err)
	}

	pterm.DefaultSection.Println("Data")
	if len(entries) > 0 {
		data := config.FormatData(existing.Values)
		err = pterm.DefaultTable.
			WithHasHeader().
			WithHeaderRowSeparator("-").
			WithData(data).
			Render()
		if err != nil {
			return fmt.Errorf("rendering data table: %w", err)
		}
	} else {
		pterm.Info.Println("This config has no keys.")
	}

	if len(existing.Workloads.Nodes) > 0 {
		pterm.DefaultSection.Println("Workloads using this config")
		return pterm.DefaultTable.
			WithHasHeader().
			WithHeaderRowSeparator("-").
			WithData(config.FormatWorkloads(existing)).
			Render()
	}

	pterm.Info.Println("No workloads are using this config.")
	return nil
}
