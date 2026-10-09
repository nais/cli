package command

import (
	"context"
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
		Examples: []naistrix.Example{
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
			{Name: "role", Prompt: "Role to assign to the team member", Choices: teamMemberRoleChoices(), ChoicesCaseInsensitive: true},
			{Name: "member", Prompt: "Email address of the team member to assign the role to"},
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
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, toComplete string) ([]string, string) {
			if args.Len() == 0 {
				return teamMemberRoleChoices(), "Choose the role to assign to the team member."
			}

			if args.Len() == 1 {
				members, err := member.GetTeamMembers(ctx, flags.Team)
				if err != nil {
					return nil, "Unable to fetch existing team members."
				}

				emails := make([]string, len(members))
				for i, m := range members {
					emails[i] = m.User.Email
				}

				return emails, "Choose the email address of the user to remove from the team."
			}

			return nil, ""
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
