package apply

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/nais/cli/internal/apply/command/flag"
	"github.com/nais/naistrix"
	"github.com/pelletier/go-toml/v2"
)

func Run(ctx context.Context, files []string, flags *flag.Apply, out naistrix.Output) error {
	a := &Apply{}

	for _, filePath := range files {
		if err := decodeFile(filePath, a); err != nil {
			return err
		}
	}

	for _, v := range a.Valkey {
		if _, err := CreateValkey(ctx, v); err != nil {
			return fmt.Errorf("failed to create valkey from file %s: %w", v.Name, err)
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
	default:
		return fmt.Errorf("unsupported file extension %s for file %s", ext, filePath)
	}
}
