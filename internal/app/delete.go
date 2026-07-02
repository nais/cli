package app

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

// DeleteApp starts deletion of an application. The Nais API deletes the
// application asynchronously, so a successful call only means deletion has been
// started.
func DeleteApp(ctx context.Context, team, name, env string) error {
	_ = `# @genqlient
		mutation DeleteApp($team: Slug!, $env: String!, $name: String!) {
		  deleteApplication(
		    input: { teamSlug: $team, environmentName: $env, name: $name }
		  ) {
		    success
		  }
		}
		`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return err
	}

	resp, err := gql.DeleteApp(ctx, client, team, env, name)
	if err != nil {
		return err
	}

	if !resp.DeleteApplication.Success {
		return fmt.Errorf("deletion of %q in %q was not successful", name, env)
	}

	return nil
}
