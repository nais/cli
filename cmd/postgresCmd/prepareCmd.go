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

func prepareCommand() *cli.Command {
	return &cli.Command{
		Name:  "prepare",
		Usage: "Prepare your postgres instance for use with personal accounts",
		Description: `Prepare will prepare the postgres instance by connecting using the
application credentials and modify the permissions on the public schema.
All IAM users in your GCP project will be able to connect to the instance.

This operation is only required to run once for each postgresql instance.`,
		ArgsUsage: "appname",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "all-privs",
				Usage: "Gives all privalges to users",
			},
		},
		Before: func(context *cli.Context) error {
			if context.Args().Len() != 1 {
				return fmt.Errorf("missing name of app")
			}

			return nil
		},
		Action: func(context *cli.Context) error {
			appName := context.Args().First()

			allPrivs := context.Bool("all-privs")
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

			return postgres.PrepareAccess(context.Context, appName, namespace, cluster, database, allPrivs)
		},
	}
}
