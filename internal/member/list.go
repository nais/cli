package member

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

func GetTeamMembers(ctx context.Context, teamSlug string) ([]gql.TeamMembersTeamMembersTeamMemberConnectionNodesTeamMember, error) {
	_ = `# @genqlient
		query TeamMembers($slug: Slug!) {
			team(slug: $slug) {
				members(first: 1000) {
					nodes {
						role
						user {
							name
							email
						}
					}
				}
			}
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := gql.TeamMembers(ctx, client, teamSlug)
	if err != nil {
		return nil, err
	}

	return resp.Team.Members.Nodes, nil
}
