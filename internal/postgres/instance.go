package postgres

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

// DeleteInstance requests removal of a Postgres and its data; cleanup continues asynchronously.
func DeleteInstance(ctx context.Context, input gql.DeletePostgresInput) (bool, error) {
	_ = `# @genqlient
		mutation DeletePostgres($input: DeletePostgresInput!) {
			deletePostgres(input: $input) { deletionRequested }
		}
	`
	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return false, err
	}
	result, err := gql.DeletePostgres(ctx, client, input)
	if err != nil {
		return false, err
	}
	return result.DeletePostgres.DeletionRequested, nil
}

// CreateInstance requests a new Postgres; provisioning continues asynchronously.
func CreateInstance(ctx context.Context, input gql.CreatePostgresInput) (string, error) {
	_ = `# @genqlient
		# @genqlient(for: "CreatePostgresInput.highAvailability", omitempty: true)
		# @genqlient(for: "CreatePostgresInput.cpu", omitempty: true)
		# @genqlient(for: "CreatePostgresInput.memory", omitempty: true)
		# @genqlient(for: "CreatePostgresInput.diskSize", omitempty: true)
		mutation CreatePostgres(
			$input: CreatePostgresInput!
		) {
			createPostgres(input: $input) { postgres { name } }
		}
	`
	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return "", err
	}
	result, err := gql.CreatePostgres(ctx, client, input)
	if err != nil {
		return "", err
	}
	return result.CreatePostgres.Postgres.Name, nil
}

// UpdateInstance requests changes to the provided Postgres fields only.
func UpdateInstance(ctx context.Context, input gql.UpdatePostgresInput) (string, error) {
	_ = `# @genqlient
		# @genqlient(for: "UpdatePostgresInput.highAvailability", omitempty: true)
		# @genqlient(for: "UpdatePostgresInput.cpu", omitempty: true)
		# @genqlient(for: "UpdatePostgresInput.memory", omitempty: true)
		# @genqlient(for: "UpdatePostgresInput.diskSize", omitempty: true)
		mutation UpdatePostgres(
			$input: UpdatePostgresInput!
		) {
			updatePostgres(input: $input) { postgres { name } }
		}
	`
	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return "", err
	}
	result, err := gql.UpdatePostgres(ctx, client, input)
	if err != nil {
		return "", err
	}
	return result.UpdatePostgres.Postgres.Name, nil
}
