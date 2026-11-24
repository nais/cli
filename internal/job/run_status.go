package job

import (
	"context"
	"fmt"
	"time"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/workload"
)

const jobRunPageSize = 100

var _ = `# @genqlient
	query FindJobRuns($team: Slug!, $env: String!, $job: String!, $first: Int!, $after: Cursor) {
		team(slug: $team) {
			environment(name: $env) {
				job(name: $job) {
					runs(first: $first, after: $after) {
						nodes { id name }
						pageInfo { hasNextPage endCursor }
					}
				}
			}
		}
	}

	query GetJobRunDetails($id: ID!, $after: Cursor) {
		node(id: $id) {
			... on JobRun {
				name
				status { state message }
				instances(first: 100, after: $after) {
					nodes {
						name
						image { name tag digest }
						restarts
						created
						status { state message lastExitReason lastExitCode lastExitTimestamp }
					}
					pageInfo { hasNextPage endCursor }
				}
			}
		}
	}
`

type RunStatus struct {
	Job         string          `json:"job"`
	Run         string          `json:"run"`
	Environment string          `json:"environment"`
	State       gql.JobRunState `json:"state"`
	Message     string          `json:"message"`
	Instances   []RunInstance   `json:"instances"`
}

type RunInstance struct {
	Name     string                  `json:"name"`
	State    gql.JobRunInstanceState `json:"state"`
	Message  string                  `json:"message"`
	Image    string                  `json:"image"`
	Restarts int                     `json:"restarts"`
	Created  time.Time               `json:"created"`
	LastExit *workload.InstanceExit  `json:"lastExit,omitempty"`
}

// RecentRunNames provides bounded completion suggestions, not the set of valid runs.
func RecentRunNames(ctx context.Context, team, env, job string) ([]string, error) {
	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := gql.FindJobRuns(ctx, client, team, env, job, jobRunPageSize, nil)
	if err != nil {
		return nil, err
	}
	runs := resp.Team.Environment.Job.Runs.Nodes
	names := make([]string, 0, len(runs))
	for _, run := range runs {
		names = append(names, run.Name)
	}
	return names, nil
}

// GetRunStatus uses the first run in the API's ordering if runName is empty.
func GetRunStatus(ctx context.Context, team, env, job, runName string) (*RunStatus, error) {
	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	// Resolve the run through the selected job; a run name alone does not establish ownership.
	latest := runName == ""
	limit := jobRunPageSize
	if latest {
		limit = 1
	}
	var id string
	var after *string
	seen := make(map[string]bool)
	for {
		resp, err := gql.FindJobRuns(ctx, client, team, env, job, limit, after)
		if err != nil {
			return nil, fmt.Errorf("finding runs for job %q in %q: %w", job, env, err)
		}
		runs := resp.Team.Environment.Job.Runs
		for _, run := range runs.Nodes {
			if latest || run.Name == runName {
				id, runName = run.Id, run.Name
				break
			}
		}
		if id != "" {
			break
		}
		if latest {
			return nil, fmt.Errorf("job %q has no runs in %q", job, env)
		}
		after, err = workload.NextCursor(runs.PageInfo.HasNextPage, runs.PageInfo.EndCursor, seen)
		if err != nil {
			return nil, fmt.Errorf("paging runs for job %q in %q: %w", job, env, err)
		}
		if after == nil {
			return nil, fmt.Errorf("run %q not found for job %q in %q", runName, job, env)
		}
	}

	ret := &RunStatus{Job: job, Run: runName, Environment: env, Instances: []RunInstance{}}
	after = nil
	clear(seen)
	for {
		resp, err := gql.GetJobRunDetails(ctx, client, id, after)
		if err != nil {
			return nil, fmt.Errorf("fetching status for run %q: %w", runName, err)
		}
		if resp.Node == nil {
			return nil, fmt.Errorf("run %q is no longer available for job %q in %q", runName, job, env)
		}
		run, ok := (*resp.Node).(*gql.GetJobRunDetailsNodeJobRun)
		if !ok || run.Name != runName {
			return nil, fmt.Errorf("run %q is no longer available for job %q in %q", runName, job, env)
		}
		if after == nil {
			ret.State, ret.Message = run.Status.State, run.Status.Message
		}
		for _, inst := range run.Instances.Nodes {
			image := inst.Image.Name
			if inst.Image.Tag != "" {
				image += ":" + inst.Image.Tag
			}
			if inst.Image.Digest != nil && *inst.Image.Digest != "" {
				image += "@" + *inst.Image.Digest
			}
			ret.Instances = append(ret.Instances, RunInstance{
				Name: inst.Name, State: inst.Status.State, Message: inst.Status.Message,
				Image: image, Restarts: inst.Restarts, Created: inst.Created,
				LastExit: workload.LastExit(inst.Status.LastExitReason, inst.Status.LastExitCode, inst.Status.LastExitTimestamp),
			})
		}
		after, err = workload.NextCursor(run.Instances.PageInfo.HasNextPage, run.Instances.PageInfo.EndCursor, seen)
		if err != nil {
			return nil, fmt.Errorf("paging instances for run %q: %w", runName, err)
		}
		if after == nil {
			return ret, nil
		}
	}
}
