package job

import (
	"context"

	"github.com/nais/cli/internal/formatting"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type Severity gql.Severity

type JobIssue struct {
	Severity    Severity `json:"severity"`
	Message     string   `json:"message"`
	Environment string   `json:"environment"`
}

func (s Severity) String() string {
	return formatting.ColoredSeverityString(string(s), gql.Severity(s))
}

func GetJobIssues(ctx context.Context, team, name string, envs []string) ([]JobIssue, error) {
	_ = `# @genqlient
		query GetJobIssues($team: Slug!, $name: String!, $env: [String!]) {
			team(slug: $team) {
				jobs(filter: { name: $name, environments: $env }) {
					nodes {
						teamEnvironment {
							environment {
								name
							}
						}
						issues(first: 500) {
							nodes {
								severity
								message
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

	resp, err := gql.GetJobIssues(ctx, client, team, name, envs)
	if err != nil {
		return nil, err
	}

	if len(resp.Team.Jobs.Nodes) == 0 {
		return nil, nil
	}

	ret := make([]JobIssue, 0)
	for _, j := range resp.Team.Jobs.Nodes {
		for _, issue := range j.Issues.Nodes {
			ret = append(ret, JobIssue{
				Severity:    Severity(issue.GetSeverity()),
				Message:     issue.GetMessage(),
				Environment: j.TeamEnvironment.Environment.Name,
			})
		}
	}

	return ret, nil
}
