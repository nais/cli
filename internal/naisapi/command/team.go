package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/cli/writer"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
)

func team(parentFlags *flag.Api) *cli.Command {
	flags := &flag.Team{
		Api:    parentFlags,
		Output: "table",
	}

	return &cli.Command{
		Name:        "team",
		Short:       "Operations on a team.",
		StickyFlags: flags,
		SubCommands: []*cli.Command{
			{
				Name:         "list-members",
				ValidateFunc: cli.ValidateExactArgs(1),
				Args: []cli.Argument{
					{Name: "team", Required: true},
				},
				RunFunc: func(ctx context.Context, out cli.Output, args []string) error {
					type member struct {
						Name  string `json:"name"`
						Email string `json:"email"`
						Role  string `json:"role"`
					}

					var members []member

					ret, err := naisapi.GetTeamMembers(ctx, args[0])
					if err != nil {
						return err
					}

					for _, m := range ret.Team.Members.Nodes {
						role := "Member"
						if m.Role == gql.TeamMemberRoleOwner {
							role = "Owner"
						}
						members = append(members, member{
							Name:  m.User.Name,
							Email: m.User.Email,
							Role:  role,
						})
					}

					if len(members) == 0 {
						out.Println("Team has no members.")
						return nil
					}

					var w writer.Writer
					if flags.Output == "json" {
						w = writer.NewJSON(out, true)
					} else {
						w = writer.NewTable(out, writer.WithColumns("Name", "Email", "Role"))
					}

					return w.Write(members)
				},
			},
		},
	}
}
