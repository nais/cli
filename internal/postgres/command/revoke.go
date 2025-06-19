package command

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"github.com/nais/cli/pkg/cli"
	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
)

func revokeCommand(parentFlags *flag.Postgres) *cli.Command {
	flags := &flag.Revoke{
		Postgres: parentFlags,
		Schema:   "public",
	}

	return &cli.Command{
		Name:  "revoke",
		Title: "Revoke access to your SQL instance for the role 'cloudsqliamuser'.",
		Description: `Revoke will revoke the role 'cloudsqliamuser' access to the tables in the SQL instance.

This is done by connecting using the application credentials and modify the permissions on the public schema.

This operation is only required to run once for each SQL instance.`,
		Args: []cli.Argument{
			{Name: "app_name", Required: true},
		},
		ValidateFunc: cli.ValidateExactArgs(1),
		Flags:        flags,
		RunFunc: func(ctx context.Context, out cli.Output, args []string) error {
			out.Println("", "Are you sure you want to continue (y/N): ")
			input := bufio.NewScanner(os.Stdin)
			input.Scan()
			if !strings.EqualFold(strings.TrimSpace(input.Text()), "y") {
				return fmt.Errorf("cancelled by user")
			}

			return postgres.RevokeAccess(ctx, args[0], flags.Namespace, flags.Context, flags.Schema)
		},
	}
}
