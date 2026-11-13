package command

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/kafka/command/flag"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func Kafka(parentFlags *flags.GlobalFlags) *naistrix.Command {
	flags := &flag.Kafka{GlobalFlags: parentFlags}
	return &naistrix.Command{
		Name:              "kafka",
		Aliases:           []string{"kafkas"},
		Title:             "Interact with Kafka topics.",
		Description:       "Commands for managing Kafka topics and credentials for your team.",
		StickyFlags:       flags,
		ValidateFlagsFunc: validation.RequireTeamFlag(flags),
		SubCommands: []*naistrix.Command{
			credentials(flags),
			grantAccess(flags),
			listGrants(flags),
			list(flags),
			revokeGrant(flags),
		},
	}
}

func validateEnvironment(flags *flag.Kafka) naistrix.ValidateFlagsFunc {
	return func(ctx context.Context) error {
		if flags.Environment == "" {
			return nil
		}
		environments, err := naisapi.GetAllEnvironments(ctx)
		if err != nil {
			return fmt.Errorf("validating Kafka environment: %w", err)
		}
		if slices.Contains(environments, string(flags.Environment)) {
			return nil
		}
		slices.Sort(environments)
		return fmt.Errorf("unknown Nais environment %q; available environments: %s", flags.Environment, strings.Join(environments, ", "))
	}
}
