package kafka

import (
	"context"
	"sort"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type Topic struct {
	Name        string `json:"name"`
	Environment string `json:"environment"`
}

func GetTeamTopics(ctx context.Context, team string, environment string, labels []gql.LabelFilter) ([]Topic, error) {
	_ = `# @genqlient
		query GetTeamKafkaTopics($team: Slug!, $filter: KafkaTopicFilter) {
			team(slug: $team) {
				kafkaTopics(first: 1000, filter: $filter) {
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

	filter := gql.KafkaTopicFilter{}
	if environment != "" {
		filter.Environments = []string{environment}
	}
	if len(labels) > 0 {
		filter.Labels = labels
	}

	resp, err := gql.GetTeamKafkaTopics(ctx, client, team, new(filter))
	if err != nil {
		return nil, err
	}

	ret := make([]Topic, 0, len(resp.Team.KafkaTopics.Nodes))
	for _, topic := range resp.Team.KafkaTopics.Nodes {
		env := topic.TeamEnvironment.Environment.Name

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

func GrantAccessToKafkaTopic(ctx context.Context, topicName, teamSlug string, environmentName flags.Environment, grant gql.KafkaTopicGrantInput) error {
	_ = `# @genqlient
		mutation GrantAccessToKafkaTopic(
			$topicName: String!
			$teamSlug: Slug!,
			$environmentName: String!,
			$grant: KafkaTopicGrantInput!,
		) {
			updateKafkaTopic(
				input: { name: $topicName, teamSlug: $teamSlug, environmentName: $environmentName, addGrants: [$grant] }
			) {
				kafkaTopic {
					id
				}
			}
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return err
	}

	if _, err = gql.GrantAccessToKafkaTopic(ctx, client, topicName, teamSlug, string(environmentName), grant); err != nil {
		return err
	}

	return nil
}
