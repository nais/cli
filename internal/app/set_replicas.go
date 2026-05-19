package app

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

func SetReplicas(ctx context.Context, team, application, env string, min, max int) (string, error) {
	_ = `# @genqlient
		mutation SetApplicationReplicas($team: Slug!, $name: String!, $env: String!, $min: Int!, $max: Int!) {
		  updateApplication(
		    input: { teamSlug: $team, environmentName: $env, name: $name, replicas: { min: $min, max: $max } }
		  ) {
		    application {
		      name
		    }
		  }
		}
			`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return "", err
	}

	resp, err := gql.SetApplicationReplicas(ctx, client, team, application, env, min, max)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Successfully updated replicas for %v in %v (min: %v, max: %v)", resp.UpdateApplication.Application.Name, env, min, max), nil
}
