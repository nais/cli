package command

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/naistrix"
)

func prepareCommand(parentFlags *flag.Postgres) *naistrix.Command {
	flags := &flag.Prepare{
		Postgres: parentFlags,
		Schema:   "public",
	}

	return &naistrix.Command{
		Name:  "prepare",
		Title: "Prepare your SQL instance for use with personal accounts.",
		Description: heredoc.Doc(`
			Prepare will prepare the SQL instance by connecting using the application credentials and modify the permissions on the public schema.

			All IAM users in your GCP project will be able to connect to the instance.

			This operation is only required to run once for each SQL instance.
		`),
		Args: []naistrix.Argument{
			{Name: "app_name"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			out.Println("", "Are you sure you want to continue (y/N): ")
			i, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}

			if !strings.EqualFold(strings.TrimSpace(i), "y") {
				return fmt.Errorf("cancelled by user")
			}

			return postgres.PrepareAccess(ctx, args.Get("app_name"), flags.Namespace, flags.Context, flags.Schema, flags.AllPrivileges)
		},
	}
}
