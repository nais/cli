package app

import (
	"context"
	"fmt"
	"time"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type ImageRelease struct {
	Image      string
	DeployedAt time.Time
}

type ApplicationImages struct {
	Current string
	History []ImageRelease
}

func GetApplicationImages(ctx context.Context, team, application, env string) (*ApplicationImages, error) {
	_ = `# @genqlient
		query GetApplicationImages($team: Slug!, $env: String!, $app: String!) {
		  team(slug: $team) {
		    environment(name: $env) {
		      application(name: $app) {
		        name
		        image {
		          name
		          tag
		        }
		        history {
		          image
		          deployedAt
		        }
		      }
		    }
		  }
		}
		`

	if team == "" {
		return nil, fmt.Errorf("team must be specified to get application images")
	}
	if application == "" {
		return nil, fmt.Errorf("application name must be specified to get application images")
	}
	if env == "" {
		return nil, fmt.Errorf("environment must be specified to get application images")
	}

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := gql.GetApplicationImages(ctx, client, team, env, application)
	if err != nil {
		return nil, err
	}

	app := resp.Team.Environment.Application
	current := app.Image.Name
	if app.Image.Tag != "" {
		current = fmt.Sprintf("%s:%s", app.Image.Name, app.Image.Tag)
	}

	history := make([]ImageRelease, 0, len(app.History))
	for _, h := range app.History {
		history = append(history, ImageRelease{
			Image:      h.Image,
			DeployedAt: h.DeployedAt,
		})
	}

	return &ApplicationImages{
		Current: current,
		History: history,
	}, nil
}

func SetImage(ctx context.Context, team, application, env, image string) (string, error) {
	_ = `# @genqlient
		mutation SetApplicationImage($team: Slug!, $name: String!, $env: String!, $image: String!) {
		  updateApplication(
		    input: { teamSlug: $team, environmentName: $env, name: $name, image: $image }
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

	resp, err := gql.SetApplicationImage(ctx, client, team, application, env, image)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Successfully set image for %v in %v to %v", resp.UpdateApplication.Application.Name, env, image), nil
}
