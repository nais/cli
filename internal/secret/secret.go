package secret

import (
	"context"
	"encoding/base64"
	"fmt"
	"slices"
	"time"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

// Metadata identifies a specific secret in a team environment.
type Metadata struct {
	// TeamSlug is the slug of the team that owns the secret.
	TeamSlug string
	// EnvironmentName is the name of the environment where the secret exists.
	EnvironmentName string
	// Name is the name of the secret.
	Name string
}

// LastModified is a time.Time that renders as human-readable relative time in
// table output (e.g. "3h", "7d") and as RFC3339 in JSON output.
type LastModified time.Time

func (t LastModified) String() string {
	v := time.Time(t)
	if v.IsZero() {
		return ""
	}

	d := time.Since(v)
	if seconds := int(d.Seconds()); seconds < -1 {
		return "<invalid>"
	} else if seconds < 0 {
		return "0s"
	} else if seconds < 60 {
		return fmt.Sprintf("%vs", seconds)
	} else if minutes := int(d.Minutes()); minutes < 60 {
		return fmt.Sprintf("%vm", minutes)
	} else if hours := int(d.Hours()); hours < 24 {
		return fmt.Sprintf("%vh", hours)
	} else if hours < 24*365 {
		return fmt.Sprintf("%vd", hours/24)
	}
	return fmt.Sprintf("%vy", int(d.Hours()/24/365))
}

func (t LastModified) MarshalJSON() ([]byte, error) {
	v := time.Time(t)
	if v.IsZero() {
		return []byte(`""`), nil
	}
	return fmt.Appendf(nil, "%q", v.Format(time.RFC3339)), nil
}

// SecretEnvironments returns the environments where a secret with the given name exists.
func SecretEnvironments(ctx context.Context, teamSlug, name string) ([]string, error) {
	all, err := GetAll(ctx, teamSlug)
	if err != nil {
		return nil, err
	}
	var envs []string
	for _, s := range all {
		if s.Name == name {
			envs = append(envs, s.TeamEnvironment.Environment.Name)
		}
	}
	return envs, nil
}

// GetAll retrieves all secrets for a team.
func GetAll(ctx context.Context, teamSlug string) ([]gql.GetAllSecretsTeamSecretsSecretConnectionNodesSecret, error) {
	_ = `# @genqlient
		query GetAllSecrets($teamSlug: Slug!) {
		  team(slug: $teamSlug) {
			secrets(first: 1000, orderBy: {field: NAME, direction: ASC}) {
			  nodes {
				name
				keys
				teamEnvironment {
				  environment {
					name
				  }
				}
				workloads(first: 1000) {
				  nodes {
					name
					__typename
				  }
				}
				lastModifiedAt
				lastModifiedBy {
				  email
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

	resp, err := gql.GetAllSecrets(ctx, client, teamSlug)
	if err != nil {
		return nil, err
	}

	return resp.Team.Secrets.Nodes, nil
}

// Get retrieves a specific secret by name in a team environment.
func Get(ctx context.Context, metadata Metadata) (*gql.GetSecretTeamEnvironmentSecret, error) {
	_ = `# @genqlient
		query GetSecret($name: String!, $environmentName: String!, $teamSlug: Slug!) {
		  team(slug: $teamSlug) {
			environment(name: $environmentName) {
			  secret(name: $name) {
				name
				keys
				teamEnvironment {
				  environment {
					name
				  }
				}
				workloads(first: 1000) {
				  nodes {
					name
					__typename
				  }
				}
				lastModifiedAt
				lastModifiedBy {
				  email
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

	resp, err := gql.GetSecret(ctx, client, metadata.Name, metadata.EnvironmentName, metadata.TeamSlug)
	if err != nil {
		return nil, err
	}

	return &resp.Team.Environment.Secret, nil
}

// Create creates a new empty secret in a team environment.
func Create(ctx context.Context, metadata Metadata) (*gql.CreateSecretCreateSecretCreateSecretPayloadSecret, error) {
	_ = `# @genqlient
		mutation CreateSecret($name: String!, $environment: String!, $team: Slug!) {
		  createSecret(input: {name: $name, environment: $environment, team: $team}) {
			secret {
			  id
			  name
			}
		  }
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := gql.CreateSecret(ctx, client, metadata.Name, metadata.EnvironmentName, metadata.TeamSlug)
	if err != nil {
		return nil, err
	}

	return &resp.CreateSecret.Secret, nil
}

// Delete deletes a secret and all its values.
func Delete(ctx context.Context, metadata Metadata) (bool, error) {
	_ = `# @genqlient
		mutation DeleteSecret($name: String!, $environment: String!, $team: Slug!) {
		  deleteSecret(input: {name: $name, environment: $environment, team: $team}) {
			secretDeleted
		  }
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return false, err
	}

	resp, err := gql.DeleteSecret(ctx, client, metadata.Name, metadata.EnvironmentName, metadata.TeamSlug)
	if err != nil {
		return false, err
	}

	return resp.DeleteSecret.SecretDeleted, nil
}

// SetValue sets a key-value pair in a secret. If the key already exists, its value is updated.
// If the key does not exist, it is added.
func SetValue(ctx context.Context, metadata Metadata, key, value string, encoding gql.ValueEncoding) (updated bool, err error) {
	existing, err := Get(ctx, metadata)
	if err != nil {
		return false, fmt.Errorf("fetching secret: %w", err)
	}

	if slices.Contains(existing.Keys, key) {
		return true, updateValue(ctx, metadata, key, value, encoding)
	}

	return false, addValue(ctx, metadata, key, value, encoding)
}

func addValue(ctx context.Context, metadata Metadata, key, value string, encoding gql.ValueEncoding) error {
	_ = `# @genqlient
		mutation AddSecretValue($name: String!, $environment: String!, $team: Slug!, $value: SecretValueInput!) {
		  addSecretValue(input: {name: $name, environment: $environment, team: $team, value: $value}) {
			secret {
			  id
			  name
			}
		  }
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return err
	}

	_, err = gql.AddSecretValue(ctx, client, metadata.Name, metadata.EnvironmentName, metadata.TeamSlug, gql.SecretValueInput{
		Name:     key,
		Value:    value,
		Encoding: encoding,
	})
	return err
}

func updateValue(ctx context.Context, metadata Metadata, key, value string, encoding gql.ValueEncoding) error {
	_ = `# @genqlient
		mutation UpdateSecretValue($name: String!, $environment: String!, $team: Slug!, $value: SecretValueInput!) {
		  updateSecretValue(input: {name: $name, environment: $environment, team: $team, value: $value}) {
			secret {
			  id
			  name
			}
		  }
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return err
	}

	_, err = gql.UpdateSecretValue(ctx, client, metadata.Name, metadata.EnvironmentName, metadata.TeamSlug, gql.SecretValueInput{
		Name:     key,
		Value:    value,
		Encoding: encoding,
	})
	return err
}

// RemoveValue removes a key-value pair from a secret.
func RemoveValue(ctx context.Context, metadata Metadata, valueName string) error {
	_ = `# @genqlient
		mutation RemoveSecretValue($secretName: String!, $environment: String!, $team: Slug!, $valueName: String!) {
		  removeSecretValue(input: {secretName: $secretName, environment: $environment, team: $team, valueName: $valueName}) {
			secret {
			  id
			  name
			}
		  }
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return err
	}

	_, err = gql.RemoveSecretValue(ctx, client, metadata.Name, metadata.EnvironmentName, metadata.TeamSlug, valueName)
	return err
}

// FormatDetails formats secret metadata for pterm table rendering.
func FormatDetails(metadata Metadata, s *gql.GetSecretTeamEnvironmentSecret) [][]string {
	data := [][]string{
		{"Field", "Value"},
		{"Team", metadata.TeamSlug},
		{"Environment", metadata.EnvironmentName},
		{"Name", s.Name},
	}

	if !s.LastModifiedAt.IsZero() {
		data = append(data, []string{"Last Modified", LastModified(s.LastModifiedAt).String()})
	}
	if s.LastModifiedBy.Email != "" {
		data = append(data, []string{"Modified By", s.LastModifiedBy.Email})
	}

	return data
}

// Entry represents a key-value pair in a secret. When values have not been
// fetched, Value is empty.
type Entry struct {
	Key      string
	Value    string
	Encoding gql.ValueEncoding
}

// FormatData formats secret keys as a key-only table for pterm rendering.
func FormatData(keys []string) [][]string {
	data := [][]string{
		{"Key"},
	}
	for _, k := range keys {
		data = append(data, []string{k})
	}
	return data
}

// FormatDataWithValues formats key-value pairs as a two-column table for pterm rendering.
// Binary values (BASE64 encoding) are shown as a placeholder with byte count.
func FormatDataWithValues(entries []Entry) [][]string {
	data := [][]string{
		{"Key", "Value"},
	}
	for _, e := range entries {
		displayValue := e.Value
		if e.Encoding == gql.ValueEncodingBase64 {
			raw, err := base64.StdEncoding.DecodeString(e.Value)
			if err == nil {
				displayValue = fmt.Sprintf("<binary, %d bytes>", len(raw))
			} else {
				displayValue = "<binary>"
			}
		}
		data = append(data, []string{e.Key, displayValue})
	}
	return data
}

// FormatWorkloads formats the workloads using a secret for pterm table rendering.
func FormatWorkloads(s *gql.GetSecretTeamEnvironmentSecret) [][]string {
	workloads := [][]string{
		{"Name", "Type"},
	}

	for _, w := range s.Workloads.Nodes {
		workloads = append(workloads, []string{
			w.GetName(),
			w.GetTypename(),
		})
	}

	return workloads
}
