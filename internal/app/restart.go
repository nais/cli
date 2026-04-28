package app

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

func RestartApp(ctx context.Context, team, application, env string) (string, error) {
	_ = `# @genqlient
		mutation RestartApp($team: Slug!, $application: String!, $env: String!) {
		  restartApplication(
		    input: { teamSlug: $team, environmentName: $env, name: $application }
		  ) {
		    application {
		      name
		    }
		  }
		}
			`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return "failed restarting application", err
	}

	if _, err := gql.RestartApp(ctx, client, team, application, env); err != nil {
		return "failed restarting application", err
	}
	return fmt.Sprintf("Successfully restarted %v in %v", application, env), nil
}
