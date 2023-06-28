package postgresCmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"github.com/nais/cli/pkg/postgres"
	"github.com/urfave/cli/v2"
)

func revokeCommand() *cli.Command {
	return &cli.Command{
		Name:  "revoke",
		Usage: "Revoke access to your postgres instance for the role 'cloudsqliamuser'",
		Description: `Revoke will revoke the role 'cloudsqliamuser' access to the
tables in the postgres instance. This is done by connecting using the application
credentials and modify the permissions on the public schema.

This operation is only required to run once for each postgresql instance.`,
		ArgsUsage: "appname",
		Before: func(context *cli.Context) error {
			if context.Args().Len() != 1 {
				return fmt.Errorf("missing name of app")
			}

			return nil
		},
		Action: func(context *cli.Context) error {
			appName := context.Args().First()

			namespace := context.String("namespace")
			cluster := context.String("context")
			database := context.String("database")

			fmt.Println(context.Command.Description)

			fmt.Print("\nAre you sure you want to continue (y/N): ")
			input := bufio.NewScanner(os.Stdin)
			input.Scan()
			if !strings.EqualFold(strings.TrimSpace(input.Text()), "y") {
				return fmt.Errorf("cancelled by user")
			}

			return postgres.RevokeAccess(context.Context, appName, namespace, cluster, database)
		},
	}
}
