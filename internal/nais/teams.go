package nais

import (
	"context"

	"github.com/nais/cli/internal/nais/gql"
)

func GetUserTeams(ctx context.Context) (*gql.UserTeamsResponse, error) {
	_ = `# @genqlient
		query UserTeams {
			me {
				... on User {
					teams {
						nodes {
							team {
								slug
								purpose
							}
						}
					}
				}
			}
		}
	`

	client, err := graphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	return gql.UserTeams(ctx, client)
}
