package apply

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/nais/cli/internal/apply/command/flag"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Run(ctx context.Context, environment, filePath string, flags *flag.Apply, out *naistrix.OutputWriter) error {
	manifests, err := loadManifests(filePath)
	if err != nil {
		return err
	}

	for _, m := range manifests {
		if m.GetNamespace() != "" {
			out.Warnf("The %v %q has namespace %q set — namespace is ignored by nais apply.\n", m.GetKind(), m.GetName(), m.GetNamespace())
		}
	}

	if err := naisapi.ApplyManifests(ctx, flags.Team, string(flags.Environment), manifests); err != nil {
		return fmt.Errorf("failed to apply manifests: %w", err)
	}

	return nil
}

// loadManifests reads all YAML documents from filePath and decodes them as CRD objects.
func loadManifests(filePath string) ([]unstructured.Unstructured, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	ext := strings.TrimLeft(filepath.Ext(filePath), ".")
	switch ext {
	case "yaml", "yml":
	default:
		return nil, fmt.Errorf("unsupported file extension %q for file %s", ext, filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	var objects []unstructured.Unstructured
	decoder := yaml.NewDecoder(bytes.NewReader(data))

	for {
		var raw unstructured.Unstructured
		if err := decoder.Decode(&raw.Object); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("failed to decode YAML from %s: %w", filePath, err)
		}
		if len(raw.Object) == 0 {
			continue
		}

		objects = append(objects, raw)
	}

	return objects, nil
}
