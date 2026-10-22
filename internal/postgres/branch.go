package postgres

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

// Branch describes the observed state of a named branch.
type Branch struct {
	Name  string                  `json:"name"`
	State gql.PostgresBranchState `json:"state"`
}

// BranchStatus distinguishes requested activation from the observed active branch.
type BranchStatus struct {
	DesiredActive string
	Active        string
	Branches      []Branch
}

func GetBranchStatus(ctx context.Context, team, environment, name string) (BranchStatus, error) {
	_ = `# @genqlient
	query GetPostgresBranchStatus($team: Slug!, $environment: String!, $postgres: String!) {
		team(slug: $team) { environment(name: $environment) { postgres(name: $postgres) {
			desiredActiveBranch activeBranch { name }
			branches(first: 1000) { nodes { name state } }
		} } }
	}
	`
	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return BranchStatus{}, err
	}
	result, err := gql.GetPostgresBranchStatus(ctx, client, team, environment, name)
	if err != nil {
		return BranchStatus{}, fmt.Errorf("fetching branches for Postgres %q: %w", name, err)
	}
	p := result.Team.Environment.Postgres
	status := BranchStatus{Branches: make([]Branch, 0, len(p.Branches.Nodes))}
	if p.DesiredActiveBranch != nil {
		status.DesiredActive = *p.DesiredActiveBranch
	}
	if p.ActiveBranch != nil {
		status.Active = p.ActiveBranch.Name
	}
	for _, branch := range p.Branches.Nodes {
		status.Branches = append(status.Branches, Branch{Name: branch.Name, State: branch.State})
	}
	return status, nil
}

type BranchWorkload struct {
	Kind string
	Name string
}

// GetBranchWorkloads lists apps/jobs selecting the branch before deletion.
func GetBranchWorkloads(ctx context.Context, team, environment, postgres, branch string) ([]BranchWorkload, error) {
	_ = `# @genqlient
	query GetPostgresBranchWorkloads($team: Slug!, $environment: String!, $postgres: String!, $branch: String!, $after: Cursor) {
		team(slug: $team) { environment(name: $environment) { postgres(name: $postgres) {
			branch(name: $branch) { workloads(first: 100, after: $after) {
				nodes { __typename name }
				pageInfo { hasNextPage endCursor }
			} }
		} } }
	}
	`
	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}
	var workloads []BranchWorkload
	var after *string
	for {
		result, err := gql.GetPostgresBranchWorkloads(ctx, client, team, environment, postgres, branch, after)
		if err != nil {
			return nil, fmt.Errorf("fetching workloads for Postgres %q branch %q before deletion: %w", postgres, branch, err)
		}
		page := result.Team.Environment.Postgres.Branch.Workloads
		for _, workload := range page.Nodes {
			if workload == nil || workload.GetTypename() == nil {
				return nil, fmt.Errorf("branch workload list contains an invalid workload")
			}
			workloads = append(workloads, BranchWorkload{Kind: *workload.GetTypename(), Name: workload.GetName()})
		}
		if !page.PageInfo.HasNextPage {
			return workloads, nil
		}
		next := page.PageInfo.EndCursor
		if next == nil || after != nil && *next == *after {
			return nil, fmt.Errorf("branch workload list has another page but no new cursor")
		}
		after = next
	}
}

func CreateBranch(ctx context.Context, input gql.CreatePostgresBranchInput) (Branch, error) {
	_ = `# @genqlient
	mutation CreatePostgresBranch($input: CreatePostgresBranchInput!) {
		createPostgresBranch(input: $input) { postgresBranch { name state } }
	}
	`
	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return Branch{}, err
	}
	result, err := gql.CreatePostgresBranch(ctx, client, input)
	if err != nil {
		return Branch{}, err
	}
	p := result.CreatePostgresBranch.PostgresBranch
	return Branch{Name: p.Name, State: p.State}, nil
}

func ActivateBranch(ctx context.Context, input gql.ActivatePostgresBranchInput) (BranchStatus, error) {
	_ = `# @genqlient
	mutation ActivatePostgresBranch($input: ActivatePostgresBranchInput!) {
		activatePostgresBranch(input: $input) { postgres { desiredActiveBranch activeBranch { name } } }
	}
	`
	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return BranchStatus{}, err
	}
	result, err := gql.ActivatePostgresBranch(ctx, client, input)
	if err != nil {
		return BranchStatus{}, err
	}
	p := result.ActivatePostgresBranch.Postgres
	status := BranchStatus{}
	if p.DesiredActiveBranch != nil {
		status.DesiredActive = *p.DesiredActiveBranch
	}
	if p.ActiveBranch != nil {
		status.Active = p.ActiveBranch.Name
	}
	return status, nil
}

func DeleteBranch(ctx context.Context, input gql.DeletePostgresBranchInput) (bool, error) {
	_ = `# @genqlient
	mutation DeletePostgresBranch($input: DeletePostgresBranchInput!) {
		deletePostgresBranch(input: $input) { postgresBranchDeleted }
	}
	`
	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return false, err
	}
	result, err := gql.DeletePostgresBranch(ctx, client, input)
	if err != nil {
		return false, err
	}
	return result.DeletePostgresBranch.PostgresBranchDeleted != nil && *result.DeletePostgresBranch.PostgresBranchDeleted, nil
}
