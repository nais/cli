package app

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type MountedFileError struct {
	Error string
}

func (e MountedFileError) String() string {
	if e.Error == "" {
		return ""
	}
	return fmt.Sprintf("<warn>%s</warn>", e.Error)
}

func (e MountedFileError) MarshalJSON() ([]byte, error) {
	if e.Error == "" {
		return []byte("null"), nil
	}
	return fmt.Appendf(nil, "%q", e.Error), nil
}

type MountedFile struct {
	Path   string           `json:"path"`
	Source ValueSource      `json:"source"`
	Error  MountedFileError `json:"error"`
}

func GetApplicationFiles(ctx context.Context, slug, name, env string) ([]MountedFile, error) {
	_ = `# @genqlient
		query GetApplicationFiles($slug: Slug!, $name: String!, $env: [String!]) {
		  team(slug: $slug) {
		    applications(filter: { name: $name, environments: $env }) {
		      nodes {
		        instanceGroups {
		          created
		          mountedFiles {
		            path
		            error
		            source {
		              kind
		              name
		            }
		          }
		        }
		      }
		    }
		  }
		}
		`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := gql.GetApplicationFiles(ctx, client, slug, name, []string{env})
	if err != nil {
		return nil, err
	}

	if len(resp.Team.Applications.Nodes) == 0 {
		return nil, fmt.Errorf("application %q not found", name)
	}

	groups := resp.Team.Applications.Nodes[0].InstanceGroups
	if len(groups) == 0 {
		return nil, nil
	}

	// Select the newest group by creation time.
	newest := groups[0]
	for _, g := range groups[1:] {
		if g.Created.After(newest.Created) {
			newest = g
		}
	}

	ret := make([]MountedFile, 0, len(newest.MountedFiles))
	for _, f := range newest.MountedFiles {
		ret = append(ret, MountedFile{
			Path: f.Path,
			Source: ValueSource{
				Kind: string(f.Source.Kind),
				Name: f.Source.Name,
			},
			Error: MountedFileError{
				Error: f.Error,
			},
		})
	}
	return ret, nil
}

// HasSecretFiles returns true if any mounted file comes from a Secret source.
func HasSecretFiles(files []MountedFile) bool {
	for _, f := range files {
		if f.Source.Kind == "SECRET" {
			return true
		}
	}
	return false
}
