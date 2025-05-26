package commands

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/cli"
)

type createFlags struct {
	*aivenFlags
	Expire uint
	Secret string
}

func create(parentFlags *aivenFlags) *cli.Command {
	createFlags := &createFlags{aivenFlags: parentFlags, Expire: 1}

	return cli.NewCommand("create", "Grant a user access to an Aiven service.",
		cli.WithValidate(func(_ context.Context, args []string) error {
			return nil
		}),
		cli.WithSubCommands(
			createKafka(createFlags),
			createOpenSearch(createFlags),
		),
		cli.WithValidate(func(_ context.Context, _ []string) error {
			if createFlags.Expire > 30 {
				return fmt.Errorf("--expire must be less than %v days", 30)
			}

			return nil
		}),
		cli.WithStickyFlag("expire", "e", "Number of `DAYS` until the generated credentials expire.", &createFlags.Expire),
		cli.WithStickyFlag("secret", "s", "`NAME` of the Kubernetes secret to store the credentials in. Will be generated if not provided.", &createFlags.Secret),
	)
}
