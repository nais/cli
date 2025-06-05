package command

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
)

func prepareCommand(parentFlags *flag.Postgres) *cli.Command {
	flags := &flag.Prepare{
		Postgres: parentFlags,
		Schema:   "public",
	}

	return cli.NewCommand("prepare", "Prepare your SQL instance for use with personal accounts.",
		cli.WithLongDescription(`Prepare will prepare the SQL instance by connecting using the
		 application credentials and modify the permissions on the public schema.
		 All IAM users in your GCP project will be able to connect to the instance.
		
		 This operation is only required to run once for each SQL instance.`),
		cli.WithArgs("app_name"),
		cli.WithValidate(cli.ValidateExactArgs(1)),
		cli.WithRun(func(ctx context.Context, out output.Output, args []string) error {
			out.Println("", "Are you sure you want to continue (y/N): ")
			i, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}

			if !strings.EqualFold(strings.TrimSpace(i), "y") {
				return fmt.Errorf("cancelled by user")
			}

			return postgres.PrepareAccess(ctx, args[0], flags.Namespace, flags.Context, flags.Schema, flags.AllPrivileges)
		}),
		cli.WithFlag("all-privs", "", "Gives all privileges to users.", &flags.AllPrivileges),
		cli.WithFlag("schema", "", "Schema to grant access to.", &flags.Schema),
	)
}
