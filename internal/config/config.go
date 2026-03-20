package config

import (
	"context"
	"encoding/base64"
	"fmt"
	"slices"
	"time"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

// Metadata identifies a specific config in a team environment.
type Metadata struct {
	// TeamSlug is the slug of the team that owns the config.
	TeamSlug string
	// EnvironmentName is the name of the environment where the config exists.
	EnvironmentName string
	// Name is the name of the config.
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

// ConfigEnvironments returns the environments where a config with the given name exists.
func ConfigEnvironments(ctx context.Context, teamSlug, name string) ([]string, error) {
	all, err := GetAll(ctx, teamSlug)
	if err != nil {
		return nil, err
	}
	var envs []string
	for _, c := range all {
		if c.Name == name {
			envs = append(envs, c.TeamEnvironment.Environment.Name)
		}
	}
	return envs, nil
}

// GetAll retrieves all configs for a team.
func GetAll(ctx context.Context, teamSlug string) ([]gql.GetAllConfigsTeamConfigsConfigConnectionNodesConfig, error) {
	_ = `# @genqlient
		query GetAllConfigs($teamSlug: Slug!) {
		  team(slug: $teamSlug) {
			configs(first: 1000, orderBy: {field: NAME, direction: ASC}) {
			  nodes {
				name
				values {
				  name
				  value
				  encoding
				}
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

	resp, err := gql.GetAllConfigs(ctx, client, teamSlug)
	if err != nil {
		return nil, err
	}

	return resp.Team.Configs.Nodes, nil
}

// Get retrieves a specific config by name in a team environment.
func Get(ctx context.Context, metadata Metadata) (*gql.GetConfigTeamEnvironmentConfig, error) {
	_ = `# @genqlient
		query GetConfig($name: String!, $environmentName: String!, $teamSlug: Slug!) {
		  team(slug: $teamSlug) {
			environment(name: $environmentName) {
			  config(name: $name) {
				name
				values {
				  name
				  value
				  encoding
				}
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

	resp, err := gql.GetConfig(ctx, client, metadata.Name, metadata.EnvironmentName, metadata.TeamSlug)
	if err != nil {
		return nil, err
	}

	return &resp.Team.Environment.Config, nil
}

// Create creates a new empty config in a team environment.
func Create(ctx context.Context, metadata Metadata) (*gql.CreateConfigCreateConfigCreateConfigPayloadConfig, error) {
	_ = `# @genqlient
		mutation CreateConfig($name: String!, $environmentName: String!, $teamSlug: Slug!) {
		  createConfig(input: {name: $name, environmentName: $environmentName, teamSlug: $teamSlug}) {
			config {
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

	resp, err := gql.CreateConfig(ctx, client, metadata.Name, metadata.EnvironmentName, metadata.TeamSlug)
	if err != nil {
		return nil, err
	}

	return &resp.CreateConfig.Config, nil
}

// Delete deletes a config and all its values.
func Delete(ctx context.Context, metadata Metadata) (bool, error) {
	_ = `# @genqlient
		mutation DeleteConfig($name: String!, $environmentName: String!, $teamSlug: Slug!) {
		  deleteConfig(input: {name: $name, environmentName: $environmentName, teamSlug: $teamSlug}) {
			configDeleted
		  }
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return false, err
	}

	resp, err := gql.DeleteConfig(ctx, client, metadata.Name, metadata.EnvironmentName, metadata.TeamSlug)
	if err != nil {
		return false, err
	}

	return resp.DeleteConfig.ConfigDeleted, nil
}

// SetValue sets a key-value pair in a config. If the key already exists, its value is updated.
// If the key does not exist, it is added.
func SetValue(ctx context.Context, metadata Metadata, key, value string, encoding gql.ValueEncoding) (updated bool, err error) {
	existing, err := Get(ctx, metadata)
	if err != nil {
		return false, fmt.Errorf("fetching config: %w", err)
	}

	keyExists := slices.ContainsFunc(existing.Values, func(v gql.GetConfigTeamEnvironmentConfigValuesConfigValue) bool {
		return v.Name == key
	})

	if keyExists {
		return true, updateValue(ctx, metadata, key, value, encoding)
	}

	return false, addValue(ctx, metadata, key, value, encoding)
}

func addValue(ctx context.Context, metadata Metadata, key, value string, encoding gql.ValueEncoding) error {
	_ = `# @genqlient
		mutation AddConfigValue($name: String!, $environmentName: String!, $teamSlug: Slug!, $value: ConfigValueInput!) {
		  addConfigValue(input: {name: $name, environmentName: $environmentName, teamSlug: $teamSlug, value: $value}) {
			config {
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

	_, err = gql.AddConfigValue(ctx, client, metadata.Name, metadata.EnvironmentName, metadata.TeamSlug, gql.ConfigValueInput{
		Name:     key,
		Value:    value,
		Encoding: encoding,
	})
	return err
}

func updateValue(ctx context.Context, metadata Metadata, key, value string, encoding gql.ValueEncoding) error {
	_ = `# @genqlient
		mutation UpdateConfigValue($name: String!, $environmentName: String!, $teamSlug: Slug!, $value: ConfigValueInput!) {
		  updateConfigValue(input: {name: $name, environmentName: $environmentName, teamSlug: $teamSlug, value: $value}) {
			config {
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

	_, err = gql.UpdateConfigValue(ctx, client, metadata.Name, metadata.EnvironmentName, metadata.TeamSlug, gql.ConfigValueInput{
		Name:     key,
		Value:    value,
		Encoding: encoding,
	})
	return err
}

// RemoveValue removes a key-value pair from a config.
func RemoveValue(ctx context.Context, metadata Metadata, valueName string) error {
	_ = `# @genqlient
		mutation RemoveConfigValue($configName: String!, $environmentName: String!, $teamSlug: Slug!, $valueName: String!) {
		  removeConfigValue(input: {configName: $configName, environmentName: $environmentName, teamSlug: $teamSlug, valueName: $valueName}) {
			config {
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

	_, err = gql.RemoveConfigValue(ctx, client, metadata.Name, metadata.EnvironmentName, metadata.TeamSlug, valueName)
	return err
}

// FormatDetails formats config metadata for pterm table rendering.
func FormatDetails(metadata Metadata, c *gql.GetConfigTeamEnvironmentConfig) [][]string {
	data := [][]string{
		{"Field", "Value"},
		{"Team", metadata.TeamSlug},
		{"Environment", metadata.EnvironmentName},
		{"Name", c.Name},
	}

	if !c.LastModifiedAt.IsZero() {
		data = append(data, []string{"Last Modified", LastModified(c.LastModifiedAt).String()})
	}
	if c.LastModifiedBy.Email != "" {
		data = append(data, []string{"Modified By", c.LastModifiedBy.Email})
	}

	return data
}

// FormatData formats config values as a key-value table for pterm rendering.
// Binary values (BASE64 encoding) are shown as a placeholder with byte count.
func FormatData(values []gql.GetConfigTeamEnvironmentConfigValuesConfigValue) [][]string {
	data := [][]string{
		{"Key", "Value"},
	}
	for _, v := range values {
		displayValue := v.Value
		if v.Encoding == gql.ValueEncodingBase64 {
			raw, err := base64.StdEncoding.DecodeString(v.Value)
			if err == nil {
				displayValue = fmt.Sprintf("<binary, %d bytes>", len(raw))
			} else {
				displayValue = "<binary>"
			}
		}
		data = append(data, []string{v.Name, displayValue})
	}
	return data
}

// FormatWorkloads formats the workloads using a config for pterm table rendering.
func FormatWorkloads(c *gql.GetConfigTeamEnvironmentConfig) [][]string {
	workloads := [][]string{
		{"Name", "Type"},
	}

	for _, w := range c.Workloads.Nodes {
		workloads = append(workloads, []string{
			w.GetName(),
			w.GetTypename(),
		})
	}

	return workloads
}
