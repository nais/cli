package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type EnvVarUpdate struct {
	Name  string
	Value *string // nil means delete
}

// ParseEnvVarUpdates parses KEY=VAL and KEY- arguments into env var updates.
// KEY=VAL sets a variable, KEY- removes it.
func ParseEnvVarUpdates(args []string) ([]EnvVarUpdate, error) {
	updates := make([]EnvVarUpdate, 0, len(args))
	for _, arg := range args {
		if strings.HasSuffix(arg, "-") && !strings.Contains(arg, "=") {
			name := strings.TrimSuffix(arg, "-")
			if name == "" {
				return nil, fmt.Errorf("invalid environment variable argument: %q", arg)
			}
			updates = append(updates, EnvVarUpdate{Name: name, Value: nil})
		} else if key, val, ok := strings.Cut(arg, "="); ok {
			if key == "" {
				return nil, fmt.Errorf("invalid environment variable argument: %q", arg)
			}
			updates = append(updates, EnvVarUpdate{Name: key, Value: &val})
		} else {
			return nil, fmt.Errorf("invalid environment variable argument: %q (expected KEY=VALUE or KEY-)", arg)
		}
	}
	return updates, nil
}

func SetApplicationEnv(ctx context.Context, team, application, env string, updates []EnvVarUpdate) (string, error) {
	_ = `# @genqlient
		# @genqlient(for: "UpdateWorkloadEnvironmentVariableInput.value", pointer: true)
		mutation SetApplicationEnv(
			$team: Slug!,
			$name: String!,
			$env: String!,
			$environmentVariables: [UpdateWorkloadEnvironmentVariableInput!]
		) {
		  updateApplication(
		    input: { teamSlug: $team, environmentName: $env, name: $name, environmentVariables: $environmentVariables }
		  ) {
		    application {
		      name
		    }
		  }
		}
			`

	envVars := make([]gql.UpdateWorkloadEnvironmentVariableInput, 0, len(updates))
	for _, u := range updates {
		envVars = append(envVars, gql.UpdateWorkloadEnvironmentVariableInput{
			Name:  u.Name,
			Value: u.Value,
		})
	}

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return "", err
	}

	resp, err := gql.SetApplicationEnv(ctx, client, team, application, env, envVars)
	if err != nil {
		return "", err
	}

	var summary []string
	for _, u := range updates {
		if u.Value == nil {
			summary = append(summary, fmt.Sprintf("  - %v (removed)", u.Name))
		} else {
			summary = append(summary, fmt.Sprintf("  + %v=%v", u.Name, *u.Value))
		}
	}

	return fmt.Sprintf("Successfully updated environment variables for %v in %v:\n%v", resp.UpdateApplication.Application.Name, env, strings.Join(summary, "\n")), nil
}
