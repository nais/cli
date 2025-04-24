package aivencmd

import (
	"github.com/nais/cli/internal/aiven"
	"github.com/urfave/cli/v2"
)

func tidyCommand() *cli.Command {
	return &cli.Command{
		Name:  "tidy",
		Usage: "Clean up /tmp/aiven-secret-* made by nais-cli",
		Description: `Remove '/tmp' folder '$TMPDIR' and files created by the aiven command
Caution - This will delete all files in '/tmp' folder starting with 'aiven-secret-'`,
		Action: func(context *cli.Context) error {
			return aiven.TidyLocalSecrets()
		},
	}
}
