package teams

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type Flags struct {
	*naisapi.Flags
}

func GetUserTeams(ctx context.Context, _ *Flags) (*gql.UserTeamsResponse, error) {
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

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	return gql.UserTeams(ctx, client)
}
