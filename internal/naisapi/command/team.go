package command

import (
	"context"
	"strings"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/cli/writer"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
	"k8s.io/utils/strings/slices"
)

func team(parentFlags *flag.Api) *cli.Command {
	flags := &flag.Team{
		Api: parentFlags,
	}

	return &cli.Command{
		Name:  "team",
		Title: "Operations on a team.",
		SubCommands: []*cli.Command{
			listMembers(flags),
			addMember(flags),
			removeMember(flags),
		},
	}
}

func listMembers(parentFlags *flag.Team) *cli.Command {
	flags := &flag.ListMembers{
		Team:   parentFlags,
		Output: "table",
	}

	return &cli.Command{
		Name:         "list-members",
		Title:        "List members of a team.",
		ValidateFunc: cli.ValidateExactArgs(1),
		Args: []cli.Argument{
			{Name: "team", Required: true},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, out cli.Output, args []string) error {
			type member struct {
				Name  string `json:"name"`
				Email string `json:"email"`
				Role  string `json:"role"`
			}

			ret, err := naisapi.GetTeamMembers(ctx, args[0])
			if err != nil {
				return err
			}

			members := make([]member, len(ret))
			for i, m := range ret {
				role := "Member"
				if m.Role == gql.TeamMemberRoleOwner {
					role = "Owner"
				}

				members[i] = member{
					Name:  m.User.Name,
					Email: m.User.Email,
					Role:  role,
				}
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
		AutoCompleteFunc: func(ctx context.Context, _ []string, toComplete string) ([]string, string) {
			if len(toComplete) < 2 {
				return nil, "Provide at least 2 characters to auto-complete team slugs."
			}

			slugs, err := naisapi.GetAllTeamSlugs(ctx)
			if err != nil {
				return nil, "Unable to fetch team slugs."
			}

			return slices.Filter([]string{}, slugs, func(slug string) bool {
				return strings.HasPrefix(slug, toComplete)
			}), "Choose a team to list members of."
		},
	}
}

func addMember(parentFlags *flag.Team) *cli.Command {
	flags := &flag.AddMember{
		Team:  parentFlags,
		Owner: false,
	}

	return &cli.Command{
		Name:         "add-member",
		Title:        "Add a member to a team.",
		Description:  "Only team owners can add team members.",
		ValidateFunc: cli.ValidateExactArgs(2),
		Examples: []cli.Example{
			{
				Description: "Add some-user@example.com to the my-team team as a regular member.",
				Command:     "my-team some-user@example.com",
			},
			{
				Description: "Add some-user@example.com to the my-team team as a team owner.",
				Command:     "my-team some-user@example.com -o",
			},
		},
		Args: []cli.Argument{
			{Name: "team", Required: true},
			{Name: "member", Required: true},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, out cli.Output, args []string) error {
			role := gql.TeamMemberRoleMember
			if flags.Owner {
				role = gql.TeamMemberRoleOwner
			}

			if err := naisapi.AddTeamMember(ctx, args[0], args[1], role); err != nil {
				return cli.Errorf("Unable to add %q to team %q:\n\n%s", args[1], args[0], err)
			}

			out.Printf("%q has been added to the %q team.\n", args[1], args[0])
			return nil
		},
		AutoCompleteFunc: func(ctx context.Context, args []string, toComplete string) ([]string, string) {
			isAdmin := naisapi.IsConsoleAdmin(ctx)

			if isAdmin && len(args) == 0 && len(toComplete) < 2 {
				return nil, "Provide at least 2 characters to auto-complete team slugs."
			}

			if len(args) == 0 {
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

func removeMember(parentFlags *flag.Team) *cli.Command {
	flags := &flag.RemoveMember{
		Team: parentFlags,
	}

	return &cli.Command{
		Name:         "remove-member",
		Title:        "Remove a member from a team.",
		Description:  "Only team owners can remove members from a team.",
		ValidateFunc: cli.ValidateExactArgs(2),
		Examples: []cli.Example{
			{
				Description: "Remove some-user@example.com from the my-team team.",
				Command:     "my-team some-user@example.com",
			},
		},
		Args: []cli.Argument{
			{Name: "team", Required: true},
			{Name: "member", Required: true},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, out cli.Output, args []string) error {
			if err := naisapi.RemoveTeamMember(ctx, args[0], args[1]); err != nil {
				return cli.Errorf("Unable to remove %q from team %q:\n\n%s", args[1], args[0], err)
			}

			out.Printf("%q has been removed from the %q team.\n", args[1], args[0])
			return nil
		},
		AutoCompleteFunc: func(ctx context.Context, args []string, toComplete string) ([]string, string) {
			isAdmin := naisapi.IsConsoleAdmin(ctx)

			if isAdmin && len(args) == 0 && len(toComplete) < 2 {
				return nil, "Provide at least 2 characters to auto-complete team slugs."
			}

			if len(args) == 0 {
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

			members, err := naisapi.GetTeamMembers(ctx, args[0])
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
