package command

import (
	"context"

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
		Examples: []naistrix.Example{
			{
				Description: "Remove some-user@example.com from the team.",
				Command:     "some-user@example.com",
			},
		},
		Args: []naistrix.Argument{
			{Name: "member"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			if err := member.RemoveTeamMember(ctx, flags.Team, args.Get("member")); err != nil {
				return naistrix.Errorf("Unable to remove %q from team %q:\n\n%s", args.Get("member"), flags.Team, err)
			}

			out.Printf("%q has been removed from the %q team.\n", args.Get("member"), flags.Team)
			return nil
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, toComplete string) ([]string, string) {
			members, err := member.GetTeamMembers(ctx, flags.Team)
			if err != nil {
				return nil, "Unable to fetch existing team members."
			}

			emails := make([]string, len(members))
			for i, m := range members {
				emails[i] = m.User.Email
			}

			return emails, "Choose the email address of the user to remove from the team."
		},
	}
}
