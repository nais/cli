package command

import (
	"context"

	"github.com/nais/cli/internal/user"
	"github.com/nais/cli/internal/user/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func list(parentFlags *flag.Users) *naistrix.Command {
	flags := &flag.ListUsers{Users: parentFlags, Output: "table"}
	return &naistrix.Command{
		Name:        "list",
		Title:       "List Nais users.",
		Description: "List Nais users with their name, email, source, and global admin status. Use <info>--source</info> to include only users from the specified sources. Without it, users with no recorded source are included. To list members of a specific team instead, use <info>nais members list --team <team></info>.",
		Flags:       flags,
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			filter, err := flags.Filter()
			if err != nil {
				return err
			}
			users, err := user.GetUsers(ctx, filter)
			if err != nil {
				return err
			}
			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(users)
			}
			if len(users) == 0 {
				out.Println("No users found.")
				return nil
			}
			return out.Table().Render(users)
		},
	}
}
