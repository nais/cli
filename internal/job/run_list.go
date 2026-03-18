package job

import (
	"context"
	"time"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type JobRun struct {
	Name     string       `json:"name"`
	Status   LastRunState `json:"status"`
	Duration string       `json:"duration"`
	Started  string       `json:"started"`
	Trigger  string       `json:"trigger"`
}

func GetJobRuns(ctx context.Context, team, environment, jobName string) ([]JobRun, error) {
	_ = `# @genqlient
		query GetJobRuns($team: Slug!, $name: String!, $env: [String!]) {
			team(slug: $team) {
				jobs(filter: { name: $name, environments: $env }, first: 1) {
					nodes {
						runs(first: 100) {
							nodes {
								name
								startTime
								duration
								status {
									state
								}
								trigger {
									type
									actor
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

	resp, err := gql.GetJobRuns(ctx, client, team, jobName, []string{environment})
	if err != nil {
		return nil, err
	}

	if len(resp.Team.Jobs.Nodes) == 0 {
		return nil, nil
	}

	runs := resp.Team.Jobs.Nodes[0].Runs.Nodes
	ret := make([]JobRun, 0, len(runs))
	for _, r := range runs {
		started := ""
		if !r.StartTime.IsZero() {
			started = r.StartTime.Format(time.DateTime)
		}

		trigger := string(r.Trigger.Type)
		if r.Trigger.Actor != "" {
			trigger = r.Trigger.Actor
		}

		ret = append(ret, JobRun{
			Name:     r.Name,
			Status:   LastRunState(r.Status.State),
			Duration: formatDuration(r.Duration),
			Started:  started,
			Trigger:  trigger,
		})
	}

	return ret, nil
}

func GetJobRunNames(ctx context.Context, team, environment string) ([]string, error) {
	_ = `# @genqlient
		query GetJobRunNames($team: Slug!) {
			team(slug: $team) {
				jobs(first: 1000) {
					nodes {
						teamEnvironment {
							environment {
								name
							}
						}
						runs(first: 100) {
							nodes {
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

	resp, err := gql.GetJobRunNames(ctx, client, team)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, j := range resp.Team.Jobs.Nodes {
		if j.TeamEnvironment.Environment.Name != environment {
			continue
		}
		for _, r := range j.Runs.Nodes {
			names = append(names, r.Name)
		}
	}
	return names, nil
}

func formatDuration(seconds int) string {
	d := time.Duration(seconds) * time.Second
	if d < time.Minute {
		return d.String()
	}
	if d < time.Hour {
		return d.Round(time.Second).String()
	}
	return d.Round(time.Minute).String()
}
