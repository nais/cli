package command

import (
	"context"
	"strings"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
	"k8s.io/utils/strings/slices"
)

func teamCommand(parentFlags *flag.Api) *naistrix.Command {
	flags := &flag.Team{Api: parentFlags}
	return &naistrix.Command{
		Name:  "team",
		Title: "Operations on a team.",
		SubCommands: []*naistrix.Command{
			listMembers(flags),
			addMember(flags),
			removeMember(flags),
			listWorkloads(flags),
		},
	}
}

func listMembers(parentFlags *flag.Team) *naistrix.Command {
	flags := &flag.ListMembers{
		Team:   parentFlags,
		Output: "table",
	}

	return &naistrix.Command{
		Name:  "list-members",
		Title: "List members of a team.",
		Args: []naistrix.Argument{
			{Name: "team"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
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

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(members)
			}

			if len(members) == 0 {
				out.Println("Team has no members.")
				return nil
			}

			return out.Table().Render(members)
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

func addMember(parentFlags *flag.Team) *naistrix.Command {
	flags := &flag.AddMember{
		Team:  parentFlags,
		Owner: false,
	}

	return &naistrix.Command{
		Name:        "add-member",
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
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			role := gql.TeamMemberRoleMember
			if flags.Owner {
				role = gql.TeamMemberRoleOwner
			}

			if err := naisapi.AddTeamMember(ctx, args[0], args[1], role); err != nil {
				return naistrix.Errorf("Unable to add %q to team %q:\n\n%s", args[1], args[0], err)
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

func removeMember(parentFlags *flag.Team) *naistrix.Command {
	flags := &flag.RemoveMember{
		Team: parentFlags,
	}

	return &naistrix.Command{
		Name:        "remove-member",
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
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			if err := naisapi.RemoveTeamMember(ctx, args[0], args[1]); err != nil {
				return naistrix.Errorf("Unable to remove %q from team %q:\n\n%s", args[1], args[0], err)
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

func listWorkloads(parentFlags *flag.Team) *naistrix.Command {
	flags := &flag.ListWorkloads{
		Team:   parentFlags,
		Output: "table",
	}

	return &naistrix.Command{
		Name:  "list-workloads",
		Title: "List workloads of a team.",
		Args: []naistrix.Argument{
			{Name: "team"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			// TODO: Once pterm/pterm#697 is resolved, we can use a link to Console instead of just the workload name.
			type workload struct {
				Name            string            `json:"name"`
				Environment     string            `json:"environment"`
				Type            string            `json:"type"`
				State           gql.WorkloadState `json:"state"`
				Vulnerabilities int               `json:"vulnerabilities"`
			}

			teamSlug := args[0]
			ret, err := naisapi.GetTeamWorkloads(ctx, teamSlug)
			if err != nil {
				return err
			}

			workloads := make([]workload, len(ret))
			for i, w := range ret {
				workloads[i] = workload{
					Name:            w.GetName(),
					Environment:     w.GetTeamEnvironment().Environment.Name,
					Type:            w.GetTypename(),
					State:           w.GetStatus().State,
					Vulnerabilities: w.GetImage().VulnerabilitySummary.Total,
				}
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(workloads)
			}

			if len(ret) == 0 {
				out.Println("Team has no workloads.")
				return nil
			}

			return out.Table().Render(workloads)
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
			}), "Choose a team to list the workloads of."
		},
	}
}
