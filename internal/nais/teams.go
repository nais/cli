package nais

import "context"

func GetUserTeams(ctx context.Context) (*UserTeamsResponse, error) {
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

	return UserTeams(ctx, client)
}
