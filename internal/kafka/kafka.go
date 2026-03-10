package kafka

import (
	"context"
	"slices"
	"sort"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type Topic struct {
	Name        string `json:"name"`
	Environment string `json:"environment"`
}

func GetTeamTopics(ctx context.Context, team string, environments []string) ([]Topic, error) {
	_ = `# @genqlient
		query GetTeamKafkaTopics($team: Slug!) {
			team(slug: $team) {
				kafkaTopics(first: 1000) {
					nodes {
						name
						teamEnvironment {
							environment {
								name
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

	resp, err := gql.GetTeamKafkaTopics(ctx, client, team)
	if err != nil {
		return nil, err
	}

	ret := make([]Topic, 0, len(resp.Team.KafkaTopics.Nodes))
	for _, topic := range resp.Team.KafkaTopics.Nodes {
		env := topic.TeamEnvironment.Environment.Name
		if len(environments) > 0 && !slices.Contains(environments, env) {
			continue
		}

		ret = append(ret, Topic{
			Name:        topic.Name,
			Environment: env,
		})
	}

	sort.Slice(ret, func(i, j int) bool {
		if ret[i].Name == ret[j].Name {
			return ret[i].Environment < ret[j].Environment
		}
		return ret[i].Name < ret[j].Name
	})

	return ret, nil
}
