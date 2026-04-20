package status

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/status/command/flag"
)

func GetStatus(ctx context.Context, _ *flag.Status) ([]gql.TeamStatusMeUserTeamsTeamMemberConnectionNodesTeamMember, error) {
	_ = `# @genqlient
		query TeamStatus {
			me {
				... on User {
					teams {
						nodes {
							team {
								slug
								workloads(first: 500) {
									nodes {
										__typename
										name
										teamEnvironment { environment { name } }
										issues(first: 100, filter: { severity: CRITICAL }) {
											nodes { __typename }
											pageInfo { totalCount }
										}
									}
									pageInfo { totalCount }
								}
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

	resp, err := gql.TeamStatus(ctx, client)
	if err != nil {
		return nil, err
	}

	if u, ok := resp.Me.(*gql.TeamStatusMeUser); ok {
		return u.Teams.Nodes, nil
	}

	return nil, nil
}
