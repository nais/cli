package app

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

func RestartApp(ctx context.Context, team, application string, envs []string) (string, error) {
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
	if len(envs) != 1 {
		return "please specify exactly one environment to restart the application in", fmt.Errorf("exactly one environment must be specified")
	}
	env := envs[0]

	_, err = gql.RestartApp(ctx, client, team, application, env)
	if err != nil {
		return "failed restarting applicatioon", err
	}
	return fmt.Sprintf("Successfully restarted %v in %v", application, env), nil
}
