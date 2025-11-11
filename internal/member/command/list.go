package command

import (
	"context"
	"strings"

	"github.com/nais/cli/internal/member"
	"github.com/nais/cli/internal/member/command/flag"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
	"k8s.io/utils/strings/slices"
)

func list(parentFlags *flag.Member) *naistrix.Command {
	flags := &flag.ListMembers{
		Member: parentFlags,
		Output: "table",
	}

	return &naistrix.Command{
		Name:  "list",
		Title: "List members of a team.",
		Flags: flags,
		ValidateFunc: func(context.Context, *naistrix.Arguments) error {
			return validation.CheckTeam(flags.Team)
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			type teamMember struct {
				Name  string `json:"name"`
				Email string `json:"email"`
				Role  string `json:"role"`
			}

			ret, err := member.GetTeamMembers(ctx, flags.Team)
			if err != nil {
				return err
			}

			members := make([]teamMember, len(ret))
			for i, m := range ret {
				role := "Member"
				if m.Role == gql.TeamMemberRoleOwner {
					role = "Owner"
				}

				members[i] = teamMember{
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
		AutoCompleteFunc: func(ctx context.Context, _ *naistrix.Arguments, toComplete string) ([]string, string) {
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
