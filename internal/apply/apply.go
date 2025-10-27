package apply

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nais/cli/internal/apply/command/flag"
	"github.com/nais/cli/internal/opensearch"
	"github.com/nais/cli/internal/valkey"
	"github.com/nais/naistrix"
	"github.com/pelletier/go-toml/v2"
)

type Apply struct {
	Version string `json:"naisVersion" toml:"naisVersion" jsonschema:"enum=v3"`
	// Valkey is a map of Valkey instances to be created, where the key is the name of the instance.
	Valkey map[string]*valkey.Valkey `json:"valkey,omitempty" toml:"valkey,omitempty"`
	// OpenSearch is a map of OpenSearch instances to be created, where the key is the name of the instance.
	OpenSearch map[string]*opensearch.OpenSearch `json:"openSearch,omitempty" toml:"openSearch,omitempty"`
}

func Run(ctx context.Context, environment, filePath string, flags *flag.Apply, out *naistrix.OutputWriter) error {
	a := &Apply{}
	if err := decodeFile(filePath, a); err != nil {
		return err
	}

	if flags.Mixin != "" {
		if err := decodeFile(string(flags.Mixin), a); err != nil {
			return fmt.Errorf("failed to decode mixin file: %w", err)
		}
	} else {
		// auto-detect mixin if not provided
		ext := filepath.Ext(filePath)
		mixinPath := strings.TrimSuffix(filePath, ext) + "." + environment + ext
		_, err := os.Stat(mixinPath)
		if err == nil {
			if flags.IsVerbose() {
				out.Println("No mixin file provided, using auto-detected mixin from " + mixinPath)
			}
			if err := decodeFile(mixinPath, a); err != nil {
				return fmt.Errorf("failed to decode mixin file: %w", err)
			}
		}
	}

	for name, v := range a.Valkey {
		metadata := valkey.Metadata{
			Name:            name,
			EnvironmentName: environment,
			TeamSlug:        flags.Team,
		}
		if err := valkey.Upsert(ctx, metadata, v); err != nil {
			return fmt.Errorf("failed to create valkey from file %s: %w", name, err)
		}
		if flags.IsVerbose() {
			out.Printf("Applied Valkey %q to environment %q for team %q\n", name, environment, flags.Team)
		}
	}

	for name, o := range a.OpenSearch {
		metadata := opensearch.Metadata{
			Name:            name,
			EnvironmentName: environment,
			TeamSlug:        flags.Team,
		}
		if err := opensearch.Upsert(ctx, metadata, o); err != nil {
			return fmt.Errorf("failed to create openSearch from file %s: %w", name, err)
		}
		if flags.IsVerbose() {
			out.Printf("Applied OpenSearch %q to environment %q for team %q\n", name, environment, flags.Team)
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
	case "toml":
		decoder := toml.NewDecoder(f)
		return decoder.DisallowUnknownFields().Decode(v)
	default:
		return fmt.Errorf("unsupported file extension %s for file %s", ext, filePath)
	}
}
