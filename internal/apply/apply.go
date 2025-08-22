package apply

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"github.com/goccy/go-yaml"
	"github.com/nais/cli/internal/apply/command/flag"
	"github.com/nais/cli/internal/opensearch"
	"github.com/nais/cli/internal/valkey"
	"github.com/nais/naistrix"
	"github.com/pelletier/go-toml/v2"
)

type Apply struct {
	Version     string `json:"naisVersion" toml:"naisVersion" jsonschema:"enum=v3"`
	Environment string `json:"environment" toml:"environment"`
	TeamSlug    string `json:"team" toml:"team"`

	// Valkey is a map of Valkey instances to be created, where the key is the name of the instance.
	Valkey map[string]*valkey.Valkey `json:"valkey,omitempty" toml:"valkey,omitempty"`
	// OpenSearch is a map of OpenSearch instances to be created, where the key is the name of the instance.
	OpenSearch map[string]*opensearch.OpenSearch `json:"openSearch,omitempty" toml:"openSearch,omitempty"`
}

func Run(ctx context.Context, files []string, _ *flag.Apply, _ naistrix.Output) error {
	a := &Apply{}

	for _, filePath := range files {
		if err := decodeFile(filePath, a); err != nil {
			return err
		}
	}

	for name, v := range a.Valkey {
		metadata := valkey.Metadata{
			Name:            name,
			EnvironmentName: a.Environment,
			TeamSlug:        a.TeamSlug,
		}
		if err := valkey.Upsert(ctx, metadata, v); err != nil {
			return fmt.Errorf("failed to create valkey from file %s: %w", name, err)
		}
	}

	for name, o := range a.OpenSearch {
		metadata := opensearch.Metadata{
			Name:            name,
			EnvironmentName: a.Environment,
			TeamSlug:        a.TeamSlug,
		}
		if err := opensearch.Upsert(ctx, metadata, o); err != nil {
			return fmt.Errorf("failed to create openSearch from file %s: %w", name, err)
		}
	}

	return nil
}

func decodeFile(filePath string, v any) error {
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	ext := strings.TrimLeft(filepath.Ext(filePath), ".")

	f, err := os.OpenFile(filePath, os.O_RDONLY, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer f.Close()

	switch ext {
	case "yaml", "yml", "json":
		decoder := yaml.NewDecoder(f, yaml.DisallowUnknownField())
		return decoder.Decode(v)
	case "toml":
		decoder := toml.NewDecoder(f)
		return decoder.DisallowUnknownFields().Decode(v)
	case "cue":
		insts := load.Instances([]string{"."}, nil)
		decoder := cuecontext.New().BuildInstance(insts[0])
		return decoder.Decode(v)
	default:
		return fmt.Errorf("unsupported file extension %s for file %s", ext, filePath)
	}
}
