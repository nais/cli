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

// SecretValue represents a key-value pair from a secret
type SecretValue struct {
	Name  string
	Value string
}

// ViewSecretValues retrieves the values of a secret. This requires team membership
// and a reason for access. The access is logged for auditing purposes.
// This is the preferred method for accessing secret values as it combines
// authorization, logging, and value retrieval in a single operation.
func ViewSecretValues(ctx context.Context, team, environmentName, secretName, reason string) ([]SecretValue, error) {
	_ = `# @genqlient
		mutation ViewSecretValues($input: ViewSecretValuesInput!) {
			viewSecretValues(input: $input) {
				values {
					name
					value
				}
			}
		}
	`

	client, err := GraphqlClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating GraphQL client: %w", err)
	}

	resp, err := gql.ViewSecretValues(ctx, client, gql.ViewSecretValuesInput{
		Name:        secretName,
		Environment: environmentName,
		Team:        team,
		Reason:      reason,
	})
	if err != nil {
		return nil, fmt.Errorf("viewing secret values: %w", err)
	}

	values := make([]SecretValue, len(resp.ViewSecretValues.Values))
	for i, v := range resp.ViewSecretValues.Values {
		values[i] = SecretValue{
			Name:  v.Name,
			Value: v.Value,
		}
	}

	return values, nil
}

// CreateElevation creates a temporary elevation of privileges for a specific resource.
// This is required before accessing sensitive resources like secrets.
// Deprecated: Use ViewSecretValues instead for accessing secret values.
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
