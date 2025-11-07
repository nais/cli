package member

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

func SetRole(ctx context.Context, teamSlug, email string, role gql.TeamMemberRole) error {
	_ = `# @genqlient
		mutation SetRole(
			$slug: Slug!
			$email: String!
			$role: TeamMemberRole!
		) {
			setTeamMemberRole(input: {teamSlug:$slug, userEmail:$email, role: $role}){
				member {
				  user {
					name
				  }
				  role
				}
		    }
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return err
	}

	_, err = gql.SetRole(ctx, client, teamSlug, email, role)
	return err
}
