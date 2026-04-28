package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func status(parentFlags *flag.App) *naistrix.Command {
	flags := &flag.Status{
		App: parentFlags,
	}

	return &naistrix.Command{
		Name:        "status",
		Title:       "Show instance status for an application.",
		Description: "Shows instance groups and their instances with current status, image, and restart counts. During rolling updates, both current and incoming groups are shown.",
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			name := args.Get("name")

			environment, err := resolveAppEnvironment(ctx, out, flags.Team, name, string(flags.Environment), flags.Output == "json")
			if err != nil {
				return err
			}

			ret, err := app.GetApplicationStatus(ctx, flags.Team, name, environment)
			if err != nil {
				return err
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(ret)
			}

			return renderStatus(out, ret)
		},
		AutoCompleteFunc: autoCompleteAppNames(flags.App),
	}
}

func renderStatus(out *naistrix.OutputWriter, status *app.InstanceGroupStatus) error {
	if len(status.Groups) == 0 {
		out.Printf("No running instances found for %s in %s.\n", status.Application, status.Environment)
		return nil
	}

	multipleGroups := len(status.Groups) > 1

	for i, group := range status.Groups {
		if i > 0 {
			out.Println("")
		}

		header := fmt.Sprintf("%s (%s)", status.Application, status.Environment)
		if multipleGroups {
			label := "Incoming"
			if group.Current {
				label = "Current"
			}
			header = fmt.Sprintf("%s — %s (%d/%d ready)", header, label, group.ReadyInstances, group.DesiredInstances)
		}
		out.Println(header)
		out.Printf("  Image: %s\n", group.Image)

		if !multipleGroups {
			out.Printf("  Instances: %d/%d ready\n", group.ReadyInstances, group.DesiredInstances)
		}
		out.Println("")

		type entry struct {
			Name     string            `json:"name"`
			Status   app.InstanceState `json:"status"`
			Restarts int               `json:"restarts"`
			Created  app.LastUpdated   `json:"created"`
		}

		entries := make([]entry, 0, len(group.Instances))
		var failingMessages []string

		for _, inst := range group.Instances {
			entries = append(entries, entry{
				Name:     inst.Name,
				Status:   inst.State,
				Restarts: inst.Restarts,
				Created:  inst.Created,
			})
			if inst.State == "FAILING" && inst.Message != "" {
				failingMessages = append(failingMessages, fmt.Sprintf("  %s: %s", inst.Name, inst.Message))
			}
		}

		if err := out.Table().Render(entries); err != nil {
			return err
		}

		if len(failingMessages) > 0 {
			out.Println("")
			out.Println(strings.Join(failingMessages, "\n"))
		}
	}
	return nil
}
