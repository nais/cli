package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/internal/member"
	"github.com/nais/cli/internal/member/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

func setRole(parentFlags *flag.Member) *naistrix.Command {
	flags := &flag.SetRole{
		Member: parentFlags,
	}

	return &naistrix.Command{
		Name:        "set-role",
		Title:       "Set role for a team member.",
		Description: "Only team owners can assign roles.",
		Examples: []naistrix.CommandExample{
			{
				Description: "Assign some-user@example.com as owner.",
				Command:     "owner some-user@example.com",
			},
			{
				Description: "Assign some-user@example.com as member.",
				Command:     "member some-user@example.com",
			},
		},
		Args: []naistrix.Argument{
			{
				Name:                   "role",
				Prompt:                 "Role to assign to the team member",
				Choices:                teamMemberRoleChoices(),
				ChoicesCaseInsensitive: true,
			},
			{
				Name: "member",
				PromptFunc: func(_ context.Context, args *naistrix.Arguments) (string, error) {
					return fmt.Sprintf("Member of %q to assign the %q role to", flags.Team, args.Get("role")), nil
				},
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
			role := gql.TeamMemberRole(strings.ToUpper(args.Get("role")))
			if err := member.SetRole(ctx, flags.Team, args.Get("member"), role); err != nil {
				return naistrix.Errorf("Unable to set role %q for member %q in team %q\n\n%s", args.Get("role"), args.Get("member"), flags.Team, err)
			}

			out.Printf("%q has been assigned %q in %q.\n", args.Get("member"), args.Get("role"), flags.Team)
			return nil
		},
	}
}

func teamMemberRoleChoices() []string {
	ret := make([]string, len(gql.AllTeamMemberRole))
	for i, role := range gql.AllTeamMemberRole {
		ret[i] = strings.ToLower(string(role))
	}
	return ret
}
