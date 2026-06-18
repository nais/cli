package apply

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

// workloadImage holds the resolved container image for a workload.
type workloadImage struct {
	name string
	tag  string
}

func (w workloadImage) String() string {
	if w.tag != "" {
		return w.name + ":" + w.tag
	}
	return w.name
}

// getWorkloadImage fetches the current container image for a workload (Application
// or Job) from the cluster via the nais API.
func getWorkloadImage(ctx context.Context, team, environment, name, kind string) (workloadImage, error) {
	switch kind {
	case "Application":
		return getApplicationImage(ctx, team, environment, name)
	case "Naisjob":
		return getJobImage(ctx, team, environment, name)
	default:
		return workloadImage{}, fmt.Errorf("unsupported workload kind %q", kind)
	}
}

func getApplicationImage(ctx context.Context, team, environment, name string) (workloadImage, error) {
	_ = `# @genqlient
		query GetCurrentApplicationImage($team: Slug!, $environment: String!, $name: String!) {
		  team(slug: $team) {
		    environment(name: $environment) {
		      application(name: $name) {
		        image {
		          name
		          tag
		        }
		      }
		    }
		  }
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return workloadImage{}, err
	}

	resp, err := gql.GetCurrentApplicationImage(ctx, client, team, environment, name)
	if err != nil {
		return workloadImage{}, err
	}

	img := resp.Team.Environment.Application.Image
	return workloadImage{name: img.Name, tag: img.Tag}, nil
}

func getJobImage(ctx context.Context, team, environment, name string) (workloadImage, error) {
	_ = `# @genqlient
		query GetCurrentJobImage($team: Slug!, $environment: String!, $name: String!) {
		  team(slug: $team) {
		    environment(name: $environment) {
		      job(name: $name) {
		        image {
		          name
		          tag
		        }
		      }
		    }
		  }
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return workloadImage{}, err
	}

	resp, err := gql.GetCurrentJobImage(ctx, client, team, environment, name)
	if err != nil {
		return workloadImage{}, err
	}

	img := resp.Team.Environment.Job.Image
	return workloadImage{name: img.Name, tag: img.Tag}, nil
}
