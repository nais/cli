package member

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

func RemoveTeamMember(ctx context.Context, teamSlug, email string) error {
	_ = `# @genqlient
		mutation RemoveTeamMember(
			$slug: Slug!
			$email: String!
		) {
			removeTeamMember(input: {
				teamSlug: $slug
				userEmail: $email
			}) {
				team { slug }
			}
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return err
	}

	_, err = gql.RemoveTeamMember(ctx, client, teamSlug, email)
	return err
}
