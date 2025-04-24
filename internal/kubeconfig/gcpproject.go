package kubeconfig

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/api/cloudresourcemanager/v3"
)

type project struct {
	ID     string
	Tenant string
	Name   string
	Kind   Kind
}

func getProjects(ctx context.Context, options filterOptions) ([]project, error) {
	var projects []project

	svc, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return nil, err
	}

	filter := "("
	filter += "(labels.naiscluster=true AND labels.environment:*)"
	filter += " OR labels.kind=legacy"
	if options.includeOnprem {
		filter += " OR labels.kind=onprem"
	}
	if options.includeManagement {
		filter += " OR labels.kind=management"
	}
	filter += ")"

	if !options.includeCi {
		filter += " AND NOT labels.environment=ci*"
	}

	if options.verbose {
		fmt.Printf("Filter: %s\n", filter)
	}

	call := svc.Projects.Search().Query(filter)
	for {
		response, err := call.Do()
		if err != nil {
			if strings.Contains(err.Error(), "invalid_grant") {
				return nil, fmt.Errorf("looks like you are missing Application Default Credentials, run `gcloud auth login --update-adc` first")
			}

			return nil, err
		}

		for _, p := range response.Projects {
			if p.State != "ACTIVE" {
				// Only check active projects. When a project is deleted,
				// it is marked as DELETING for a while before it is removed.
				// This results in a 403 when trying to list clusters.
				continue
			}

			projects = append(projects, project{
				ID:     p.ProjectId,
				Tenant: p.Labels["tenant"],
				Name:   p.Labels["environment"],
				Kind:   parseKind(p.Labels["kind"]),
			})
		}
		if response.NextPageToken == "" {
			break
		}
		call.PageToken(response.NextPageToken)
	}
	if options.verbose {
		fmt.Printf("Projects:\n")
		for _, p := range projects {
			fmt.Printf("%s\t%s\t%s\t%v\n", p.ID, p.Tenant, p.Name, p.Kind)
		}
	}

	return projects, nil
}
