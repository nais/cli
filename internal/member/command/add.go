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

func add(parentFlags *flag.Member) *naistrix.Command {
	flags := &flag.AddMember{
		Member: parentFlags,
		Owner:  false,
	}

	return &naistrix.Command{
		Name:        "add",
		Title:       "Add a member to a team.",
		Description: "Only team owners can add team members.",
		Examples: []naistrix.Example{
			{
				Description: "Add some-user@example.com to the my-team team as a regular member.",
				Command:     "my-team some-user@example.com",
			},
			{
				Description: "Add some-user@example.com to the my-team team as a team owner.",
				Command:     "my-team some-user@example.com -o",
			},
		},
		Args: []naistrix.Argument{
			{Name: "team"},
			{Name: "member"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			role := gql.TeamMemberRoleMember
			if flags.Owner {
				role = gql.TeamMemberRoleOwner
			}

			if err := member.AddTeamMember(ctx, args.Get("team"), args.Get("member"), role); err != nil {
				return naistrix.Errorf("Unable to add %q to team %q:\n\n%s", args.Get("member"), args.Get("team"), err)
			}

			out.Printf("%q has been added to the %q team.\n", args.Get("member"), args.Get("team"))
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

				return slugs, "Choose a team to add a member to."
			}

			if len(toComplete) < 2 {
				return nil, "Provide at least 2 characters to auto-complete user emails."
			}

			emails, err := naisapi.GetUserEmails(ctx)
			if err != nil {
				return nil, "Unable to fetch user emails."
			}

			return slices.Filter([]string{}, emails, func(email string) bool {
				return strings.HasPrefix(email, toComplete)
			}), "Choose the email address of the user to add to the team."
		},
	}
}
