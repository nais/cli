package resource

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/config"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

func init() {
	register(configResource{kindSupport{kind: "Config", strippedVersion: "v1"}})
}

// configResource applies Config instances through the nais-api Config mutations.
// Unlike Valkey and OpenSearch, Config uses top-level data/binaryData fields
// instead of spec, and the API operates on individual values rather than a
// single create-or-update mutation.
type configResource struct{ kindSupport }

func (c configResource) Apply(ctx context.Context, meta Metadata, m Manifest) (Action, error) {
	cmeta := config.Metadata{
		Name:            meta.Name,
		TeamSlug:        meta.TeamSlug,
		EnvironmentName: meta.EnvironmentName,
	}

	existing, action, err := c.ensureExists(ctx, cmeta)
	if err != nil {
		return "", err
	}

	// Build a set of existing values for efficient lookup.
	existingValues := make(map[string]gql.GetConfigTeamEnvironmentConfigValuesConfigValue, len(existing.Values))
	for _, v := range existing.Values {
		existingValues[v.Name] = v
	}

	// Desired state: merge data (PLAIN_TEXT) and binaryData (BASE64).
	desired := make(map[string]configValue, len(m.Data)+len(m.BinaryData))
	for k, v := range m.Data {
		desired[k] = configValue{value: v, encoding: gql.ValueEncodingPlainText}
	}
	for k, v := range m.BinaryData {
		desired[k] = configValue{value: v, encoding: gql.ValueEncodingBase64}
	}

	// Add or update values present in the manifest.
	for key, dv := range desired {
		if ev, exists := existingValues[key]; exists {
			if ev.Value == dv.value && ev.Encoding == dv.encoding {
				continue // unchanged
			}
			if _, err := config.SetValue(ctx, cmeta, key, dv.value, dv.encoding); err != nil {
				return "", fmt.Errorf("updating value %q: %w", key, err)
			}
			action = ActionUpdated
		} else {
			if _, err := config.SetValue(ctx, cmeta, key, dv.value, dv.encoding); err != nil {
				return "", fmt.Errorf("adding value %q: %w", key, err)
			}
			if action != ActionCreated {
				action = ActionUpdated
			}
		}
	}

	// Remove values not present in the manifest.
	for key := range existingValues {
		if _, keep := desired[key]; !keep {
			if err := config.RemoveValue(ctx, cmeta, key); err != nil {
				return "", fmt.Errorf("removing value %q: %w", key, err)
			}
			action = ActionUpdated
		}
	}

	return action, nil
}

type configValue struct {
	value    string
	encoding gql.ValueEncoding
}

// ensureExists creates the config if it does not exist, returning the current
// state and whether it was just created.
func (c configResource) ensureExists(ctx context.Context, meta config.Metadata) (*gql.GetConfigTeamEnvironmentConfig, Action, error) {
	existing, err := config.Get(ctx, meta)
	if err == nil {
		return existing, ActionUpdated, nil
	}

	if !naisapi.IsNotFound(err) {
		return nil, "", err
	}

	if _, err := config.Create(ctx, meta); err != nil {
		return nil, "", fmt.Errorf("creating config: %w", err)
	}

	// Fetch the newly created config to get a consistent state.
	existing, err = config.Get(ctx, meta)
	if err != nil {
		return nil, "", fmt.Errorf("fetching newly created config: %w", err)
	}

	return existing, ActionCreated, nil
}
