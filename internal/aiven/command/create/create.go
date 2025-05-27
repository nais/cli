package create

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/aiven/flag"
	"github.com/nais/cli/internal/cli"
)

func Create(parentFlags *flag.Aiven) *cli.Command {
	createFlags := &flag.Create{Aiven: parentFlags, Expire: 1}

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
