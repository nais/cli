package command

import (
	"context"

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
				Description: "Assign some-user@example.com OWNER in my-team.",
				Command:     "my-team OWNER some-user@example.com",
			},
			{
				Description: "Assign some-user@example.com MEMBER in my-team.",
				Command:     "my-team MEMBER some-user@example.com",
			},
		},
		Args: []naistrix.Argument{
			{Name: "team"},
			{Name: "role"},
			{Name: "member"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			if err := member.SetRole(ctx, args.Get("team"), args.Get("member"), gql.TeamMemberRole(args.Get("role"))); err != nil {
				return naistrix.Errorf("Unable to set role %q for member %q in team %q\n\n%s", args.Get("role"), args.Get("member"), args.Get("team"), err)
			}

			out.Printf("%q has been assigned %q in %q.\n", args.Get("member"), args.Get("role"), args.Get("team"))
			return nil
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, toComplete string) ([]string, string) {
			if args.Len() == 1 {
				return toStrings(gql.AllTeamMemberRole), "Choose the role to assign to the team member."
			}

			if args.Len() == 2 {
				members, err := member.GetTeamMembers(ctx, args.Get("team"))
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

func toStrings[T ~string](in []T) []string {
	ret := make([]string, len(in))
	for i, s := range in {
		ret[i] = string(s)
	}
	return ret
}
