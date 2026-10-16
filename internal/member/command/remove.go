package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/member"
	"github.com/nais/cli/internal/member/command/flag"
	"github.com/nais/naistrix"
)

func remove(parentFlags *flag.Member) *naistrix.Command {
	flags := &flag.RemoveMember{
		Member: parentFlags,
	}

	return &naistrix.Command{
		Name:        "remove",
		Title:       "Remove a member from a team.",
		Description: "Only team owners can remove members from a team.",
		Examples: []naistrix.CommandExample{
			{
				Description: "Remove some-user@example.com from the team.",
				Command:     "some-user@example.com",
			},
		},
		Args: []naistrix.Argument{
			{
				Name:   "member",
				Prompt: "Email address of the user to remove from the team",
				ChoicesFunc: func(ctx context.Context, _ *naistrix.Arguments) ([]string, error) {
					members, err := member.GetTeamMembers(ctx, flags.Team)
					if err != nil {
						return nil, fmt.Errorf("unable to fetch existing team members: %w", err)
					}

					emails := make([]string, len(members))
					for i, m := range members {
						emails[i] = m.User.Email
					}
					return emails, nil
				},
			},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			if err := member.RemoveTeamMember(ctx, flags.Team, args.Get("member")); err != nil {
				return naistrix.Errorf("Unable to remove %q from team %q:\n\n%s", args.Get("member"), flags.Team, err)
			}

			out.Printf("%q has been removed from the %q team.\n", args.Get("member"), flags.Team)
			return nil
		},
	}
}
