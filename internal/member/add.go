package member

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

func AddTeamMember(ctx context.Context, teamSlug, email string, role gql.TeamMemberRole) error {
	_ = `# @genqlient
		mutation AddTeamMember(
			$slug: Slug!
			$email: String!
			$role: TeamMemberRole!
		) {
			addTeamMember(input: {
				teamSlug: $slug
				userEmail: $email
				role: $role
			}) {
				member { role }
			}
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return err
	}

	_, err = gql.AddTeamMember(ctx, client, teamSlug, email, role)
	return err
}
