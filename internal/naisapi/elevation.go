package naisapi

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi/gql"
)

// ElevationType represents the type of elevation to request
type ElevationType = gql.ElevationType

const (
	ElevationTypeSecret ElevationType = gql.ElevationTypeSecret
)

// Elevation represents an active elevation
type Elevation struct {
	ID              string
	Type            ElevationType
	TeamSlug        string
	EnvironmentName string
	ResourceName    string
	Reason          string
}

// CreateElevation creates a temporary elevation of privileges for a specific resource.
// This is required before accessing sensitive resources like secrets.
func CreateElevation(ctx context.Context, team, environmentName, resourceName, reason string, durationMinutes int) (*Elevation, error) {
	if durationMinutes <= 0 {
		durationMinutes = 5
	}

	_ = `# @genqlient
		mutation CreateElevation($input: CreateElevationInput!) {
			createElevation(input: $input) {
				elevation {
					id
					type
					team {
						slug
					}
					resourceName
					reason
				}
			}
		}
	`

	client, err := GraphqlClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating GraphQL client: %w", err)
	}

	resp, err := gql.CreateElevation(ctx, client, gql.CreateElevationInput{
		Type:            gql.ElevationTypeSecret,
		Team:            team,
		EnvironmentName: environmentName,
		ResourceName:    resourceName,
		Reason:          reason,
		DurationMinutes: durationMinutes,
	})
	if err != nil {
		return nil, fmt.Errorf("creating elevation: %w", err)
	}

	elev := resp.CreateElevation.Elevation
	return &Elevation{
		ID:              elev.Id,
		Type:            elev.Type,
		TeamSlug:        elev.Team.Slug,
		EnvironmentName: environmentName,
		ResourceName:    elev.ResourceName,
		Reason:          elev.Reason,
	}, nil
}
