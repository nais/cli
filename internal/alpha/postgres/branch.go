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

// GetNamedBranchStatus returns the requested branch's state and its Postgres activation state.
func GetNamedBranchStatus(ctx context.Context, team, environment, name, branch string) (BranchStatus, error) {
	_ = `# @genqlient
	query GetNamedPostgresBranchStatus($team: Slug!, $environment: String!, $postgres: String!, $branch: String!) {
		team(slug: $team) { environment(name: $environment) { postgres(name: $postgres) {
			desiredActiveBranch activeBranch { name }
			branch(name: $branch) { name state }
		} } }
	}
	`
	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return BranchStatus{}, err
	}
	result, err := gql.GetNamedPostgresBranchStatus(ctx, client, team, environment, name, branch)
	if err != nil {
		return BranchStatus{}, fmt.Errorf("fetching branch %q of Postgres %q: %w", branch, name, err)
	}
	p := result.Team.Environment.Postgres
	status := BranchStatus{Branches: []Branch{{Name: p.Branch.Name, State: p.Branch.State}}}
	if p.DesiredActiveBranch != nil {
		status.DesiredActive = *p.DesiredActiveBranch
	}
	if p.ActiveBranch != nil {
		status.Active = p.ActiveBranch.Name
	}
	return status, nil
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
