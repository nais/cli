package command

import (
	"context"
	"strings"

	"github.com/nais/cli/internal/member"
	"github.com/nais/cli/internal/member/command/flag"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
	"k8s.io/utils/strings/slices"
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
				Description: "Remove some-user@example.com from the my-team team.",
				Command:     "my-team some-user@example.com",
			},
		},
		Args: []naistrix.Argument{
			{Name: "team"},
			{Name: "member"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			if err := member.RemoveTeamMember(ctx, args.Get("team"), args.Get("member")); err != nil {
				return naistrix.Errorf("Unable to remove %q from team %q:\n\n%s", args.Get("member"), args.Get("team"), err)
			}

			out.Printf("%q has been removed from the %q team.\n", args.Get("member"), args.Get("team"))
			return nil
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, toComplete string) ([]string, string) {
			isAdmin := naisapi.IsConsoleAdmin(ctx)

			if isAdmin && args.Len() == 0 && len(toComplete) < 2 {
				return nil, "Provide at least 2 characters to auto-complete team slugs."
			}

			if args.Len() == 0 {
				var slugs []string

				if isAdmin {
					allSlugs, err := naisapi.GetAllTeamSlugs(ctx)
					if err != nil {
						return nil, "Unable to fetch team slugs."
					}

					slugs = slices.Filter([]string{}, allSlugs, func(slug string) bool {
						return strings.HasPrefix(slug, toComplete)
					})
				} else {
					userTeams, err := naisapi.GetUserTeams(ctx)
					if err != nil {
						return nil, "Unable to fetch team slugs."
					}

					for _, t := range userTeams {
						if t.Role == gql.TeamMemberRoleOwner {
							slugs = append(slugs, t.Team.Slug)
						}
					}
				}

				if len(slugs) == 0 {
					return nil, "You are not an owner of any teams."
				}

				return slugs, "Choose a team to remove a member from."
			}

			members, err := member.GetTeamMembers(ctx, args.Get("team"))
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
