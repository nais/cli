package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
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
		Description: "Shows readiness, instance groups, stuck rollout warnings, and observed automatic scaling for an application.",
		Args: []naistrix.Argument{
			{Name: "name", Prompt: "Name of the application to show instance status for", ChoicesFunc: appNameChoices(parentFlags)},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			name := args.Get("name")

			environment, err := resolveAppEnvironment(ctx, flags.Team, name, string(flags.Environment))
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
	}
}

func renderStatus(out *naistrix.OutputWriter, status *app.InstanceGroupStatus) error {
	header := fmt.Sprintf("%s (%s)", status.Application, status.Environment)
	out.Println(header)
	if rollout := status.Rollout; rollout != nil {
		switch rollout.Reason {
		case gql.ApplicationRolloutStuckReasonInstanceCreationFailed:
			out.Warnf("Rollout stuck — instance creation failed: %s\n", rollout.Message)
			out.Println("  Check nais app events for details.")
		case gql.ApplicationRolloutStuckReasonTimedOut:
			out.Warnf("Rollout stuck — timed out: %s\n", rollout.Message)
			if len(status.Groups) > 1 {
				out.Println("  Check the newest instance group below.")
			}
		default:
			out.Warnf("Rollout stuck: %s\n", rollout.Message)
		}
	}
	if scaling := status.Scaling; scaling != nil {
		out.Printf("Automatic scaling: %s (%d current, %d desired)\n", scalingLabel(scaling.State), scaling.CurrentInstances, scaling.DesiredInstances)
		if scaling.Message != "" {
			out.Printf("  %s\n", scaling.Message)
		}
	}
	if len(status.Groups) == 0 {
		out.Println("No instance groups observed.")
		return nil
	}

	multipleGroups := len(status.Groups) > 1

	for i, group := range status.Groups {
		out.Println("")
		if multipleGroups {
			label := "Earlier"
			if i == 0 {
				label = "Newest"
			}
			out.Printf("%s group %s (%d/%d ready)\n", label, group.Name, group.ReadyInstances, group.DesiredInstances)
		} else {
			out.Printf("  Instances: %d/%d ready\n", group.ReadyInstances, group.DesiredInstances)
		}
		out.Printf("  Image: %s\n", group.Image)

		out.Println("")

		type entry struct {
			Name     string          `json:"name"`
			Status   string          `json:"status"`
			Restarts int             `json:"restarts"`
			Created  app.LastUpdated `json:"created"`
		}

		entries := make([]entry, 0, len(group.Instances))
		var details []string

		for _, inst := range group.Instances {
			label := inst.State.String()
			if inst.State == app.InstanceState(gql.ApplicationInstanceStateRunning) && !inst.Ready {
				label = "Running, not ready"
			}
			entries = append(entries, entry{
				Name:     inst.Name,
				Status:   label,
				Restarts: inst.Restarts,
				Created:  inst.Created,
			})
			if !inst.Ready && inst.Message != "" {
				details = append(details, fmt.Sprintf("  %s: %s", inst.Name, inst.Message))
			}
			if inst.LastExit != nil && (inst.Restarts > 0 || !inst.Ready) {
				details = append(details, fmt.Sprintf("  %s last exit: %s", inst.Name, inst.LastExit))
			}
		}

		if len(entries) > 0 {
			if err := out.Table().Render(entries); err != nil {
				return err
			}
		}

		if len(details) > 0 {
			out.Println("")
			out.Println(strings.Join(details, "\n"))
		}
	}
	return nil
}

func scalingLabel(state gql.ApplicationScalingState) string {
	switch state {
	case gql.ApplicationScalingStateActive:
		return "Active"
	case gql.ApplicationScalingStateLimitedByMaximum:
		return "Limited by configured maximum"
	case gql.ApplicationScalingStateMetricsUnavailable:
		return "Metrics unavailable"
	case gql.ApplicationScalingStateUnableToScale:
		return "Unable to scale"
	case gql.ApplicationScalingStateDisabled:
		return "Disabled"
	default:
		return string(state)
	}
}
