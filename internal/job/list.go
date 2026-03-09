package job

import (
	"context"
	"slices"
	"sort"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type (
	State        string
	LastRunState string
	Schedule     string
)

type Job struct {
	Name        string       `json:"name"`
	Environment string       `json:"environment"`
	Schedule    Schedule     `json:"schedule"`
	LastRun     LastRunState `heading:"Last Run" json:"last_run"`
	State       State        `json:"state"`
	Issues      int          `json:"issues"`
}

func (s State) String() string {
	switch s {
	case State(gql.JobStateRunning):
		return "Running"
	case State(gql.JobStateCompleted):
		return "Completed"
	case State(gql.JobStateFailed):
		return "<error>Failed</error>"
	default:
		return "<info>Unknown</info>"
	}
}

func (s LastRunState) String() string {
	switch s {
	case LastRunState(gql.JobRunStateRunning):
		return "Running"
	case LastRunState(gql.JobRunStateSucceeded):
		return "Succeeded"
	case LastRunState(gql.JobRunStateFailed):
		return "<error>Failed</error>"
	case LastRunState(gql.JobRunStatePending):
		return "Pending"
	default:
		return "<unknown>"
	}
}

func (s Schedule) String() string {
	if s == "" {
		return "Once"
	}
	return string(s)
}

func GetTeamJobs(ctx context.Context, team string, environments []string) ([]Job, error) {
	_ = `# @genqlient
		query GetTeamJobs($team: Slug!, $orderBy: JobOrder) {
			team(slug: $team) {
				jobs(first: 1000, orderBy: $orderBy) {
					nodes {
						name
						teamEnvironment {
							environment {
								name
							}
						}
						state
						schedule {
							expression
						}
						runs(first: 1) {
							nodes {
								status {
									state
								}
							}
						}
						issues {
							pageInfo {
								totalCount
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

	resp, err := gql.GetTeamJobs(ctx, client, team, gql.JobOrder{
		Field:     gql.JobOrderFieldIssues,
		Direction: gql.OrderDirectionDesc,
	})
	if err != nil {
		return nil, err
	}

	ret := make([]Job, 0)
	for _, j := range resp.Team.Jobs.Nodes {
		env := j.TeamEnvironment.Environment.Name
		if len(environments) > 0 && !slices.Contains(environments, env) {
			continue
		}

		schedule := Schedule("")
		if j.Schedule.Expression != "" {
			schedule = Schedule(j.Schedule.Expression)
		}

		lastRun := LastRunState("")
		if len(j.Runs.GetNodes()) > 0 {
			lastRun = LastRunState(j.Runs.GetNodes()[0].Status.State)
		}

		ret = append(ret, Job{
			Name:        j.Name,
			Environment: env,
			Schedule:    schedule,
			LastRun:     lastRun,
			State:       State(j.State),
			Issues:      j.Issues.PageInfo.TotalCount,
		})
	}

	sort.Slice(ret, func(i, j int) bool {
		if ret[i].Issues == ret[j].Issues {
			return ret[i].Name < ret[j].Name
		}
		return ret[i].Issues > ret[j].Issues
	})

	return ret, nil
}
