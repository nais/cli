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
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/opensearch"
	"github.com/nais/cli/internal/valkey"
	"github.com/nais/naistrix"
	pgratorv1 "github.com/nais/pgrator/pkg/api/v1"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	scheme = runtime.NewScheme()
	_      = pgratorv1.AddToScheme(scheme)
	codecs = serializer.NewCodecFactory(scheme)
)

func Run(ctx context.Context, environment, filePath string, flags *flag.Apply, out *naistrix.OutputWriter) error {
	manifests, err := loadManifests(filePath)
	if err != nil {
		return err
	}

	for _, m := range manifests {
		switch obj := m.(type) {
		case *pgratorv1.Valkey:
			if obj.Namespace != "" {
				out.Warnf("Valkey %q has namespace %q set — namespace is ignored by nais apply.\n", obj.Name, obj.Namespace)
			}
			metadata := valkey.Metadata{
				Name:            obj.Name,
				EnvironmentName: environment,
				TeamSlug:        flags.Team,
			}
			v, err := valkeyFromCRD(obj)
			if err != nil {
				return fmt.Errorf("valkey %q: %w", obj.Name, err)
			}
			if err := valkey.Upsert(ctx, metadata, v); err != nil {
				return fmt.Errorf("failed to apply Valkey %q: %w", obj.Name, err)
			}
			if flags.IsVerbose() {
				out.Printf("Applied Valkey %q to environment %q for team %q\n", obj.Name, environment, flags.Team)
			}

		case *pgratorv1.OpenSearch:
			if obj.Namespace != "" {
				out.Warnf("OpenSearch %q has namespace %q set — namespace is ignored by nais apply.\n", obj.Name, obj.Namespace)
			}
			metadata := opensearch.Metadata{
				Name:            obj.Name,
				EnvironmentName: environment,
				TeamSlug:        flags.Team,
			}
			o, err := openSearchFromCRD(obj)
			if err != nil {
				return fmt.Errorf("openSearch %q: %w", obj.Name, err)
			}
			if err := opensearch.Upsert(ctx, metadata, o); err != nil {
				return fmt.Errorf("failed to apply OpenSearch %q: %w", obj.Name, err)
			}
			if flags.IsVerbose() {
				out.Printf("Applied OpenSearch %q to environment %q for team %q\n", obj.Name, environment, flags.Team)
			}

		default:
			return fmt.Errorf("unsupported resource type %T", m)
		}
	}

	return nil
}

// loadManifests reads all YAML documents from filePath and decodes them as CRD objects.
func loadManifests(filePath string) ([]runtime.Object, error) {
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

	var objects []runtime.Object
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	deserializer := codecs.UniversalDeserializer()

	for {
		var raw map[string]any
		if err := decoder.Decode(&raw); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("failed to decode YAML from %s: %w", filePath, err)
		}
		if len(raw) == 0 {
			continue
		}

		// Re-encode the single document back to YAML bytes for the k8s deserializer.
		docBytes, err := yaml.Marshal(raw)
		if err != nil {
			return nil, fmt.Errorf("failed to re-encode YAML document from %s: %w", filePath, err)
		}

		obj, _, err := deserializer.Decode(docBytes, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to decode manifest from %s: %w", filePath, err)
		}

		objects = append(objects, obj)
	}

	return objects, nil
}

var valkeyTierMap = map[pgratorv1.ValkeyTier]gql.ValkeyTier{
	pgratorv1.ValkeyTierSingleNode:       gql.ValkeyTierSingleNode,
	pgratorv1.ValkeyTierHighAvailability: gql.ValkeyTierHighAvailability,
}

var valkeyMemoryMap = map[pgratorv1.ValkeyMemory]gql.ValkeyMemory{
	pgratorv1.ValkeyMemory1GB:   gql.ValkeyMemoryGb1,
	pgratorv1.ValkeyMemory4GB:   gql.ValkeyMemoryGb4,
	pgratorv1.ValkeyMemory8GB:   gql.ValkeyMemoryGb8,
	pgratorv1.ValkeyMemory14GB:  gql.ValkeyMemoryGb14,
	pgratorv1.ValkeyMemory28GB:  gql.ValkeyMemoryGb28,
	pgratorv1.ValkeyMemory56GB:  gql.ValkeyMemoryGb56,
	pgratorv1.ValkeyMemory112GB: gql.ValkeyMemoryGb112,
	pgratorv1.ValkeyMemory200GB: gql.ValkeyMemoryGb200,
}

var valkeyMaxMemoryPolicyMap = map[pgratorv1.ValkeyMaxMemoryPolicy]gql.ValkeyMaxMemoryPolicy{
	pgratorv1.ValkeyMaxMemoryPolicyAllkeysLFU:     gql.ValkeyMaxMemoryPolicyAllkeysLfu,
	pgratorv1.ValkeyMaxMemoryPolicyAllkeysLRU:     gql.ValkeyMaxMemoryPolicyAllkeysLru,
	pgratorv1.ValkeyMaxMemoryPolicyAllkeysRandom:  gql.ValkeyMaxMemoryPolicyAllkeysRandom,
	pgratorv1.ValkeyMaxMemoryPolicyNoEviction:     gql.ValkeyMaxMemoryPolicyNoEviction,
	pgratorv1.ValkeyMaxMemoryPolicyVolatileLFU:    gql.ValkeyMaxMemoryPolicyVolatileLfu,
	pgratorv1.ValkeyMaxMemoryPolicyVolatileLRU:    gql.ValkeyMaxMemoryPolicyVolatileLru,
	pgratorv1.ValkeyMaxMemoryPolicyVolatileRandom: gql.ValkeyMaxMemoryPolicyVolatileRandom,
	pgratorv1.ValkeyMaxMemoryPolicyVolatileTTL:    gql.ValkeyMaxMemoryPolicyVolatileTtl,
}

func valkeyFromCRD(obj *pgratorv1.Valkey) (*valkey.Valkey, error) {
	tier, ok := valkeyTierMap[obj.Spec.Tier]
	if !ok {
		return nil, fmt.Errorf("unsupported tier %q", obj.Spec.Tier)
	}

	mem, ok := valkeyMemoryMap[obj.Spec.Memory]
	if !ok {
		return nil, fmt.Errorf("unsupported memory %q", obj.Spec.Memory)
	}

	var maxMemPolicy gql.ValkeyMaxMemoryPolicy
	if obj.Spec.MaxMemoryPolicy != "" {
		maxMemPolicy, ok = valkeyMaxMemoryPolicyMap[obj.Spec.MaxMemoryPolicy]
		if !ok {
			return nil, fmt.Errorf("unsupported maxMemoryPolicy %q", obj.Spec.MaxMemoryPolicy)
		}
	}

	return &valkey.Valkey{
		Tier:            tier,
		Memory:          mem,
		MaxMemoryPolicy: maxMemPolicy,
	}, nil
}

var openSearchTierMap = map[pgratorv1.OpenSearchTier]gql.OpenSearchTier{
	pgratorv1.OpenSearchTierSingleNode:       gql.OpenSearchTierSingleNode,
	pgratorv1.OpenSearchTierHighAvailability: gql.OpenSearchTierHighAvailability,
}

var openSearchMemoryMap = map[pgratorv1.OpenSearchMemory]gql.OpenSearchMemory{
	pgratorv1.OpenSearchMemory2GB:  gql.OpenSearchMemoryGb2,
	pgratorv1.OpenSearchMemory4GB:  gql.OpenSearchMemoryGb4,
	pgratorv1.OpenSearchMemory8GB:  gql.OpenSearchMemoryGb8,
	pgratorv1.OpenSearchMemory16GB: gql.OpenSearchMemoryGb16,
	pgratorv1.OpenSearchMemory32GB: gql.OpenSearchMemoryGb32,
	pgratorv1.OpenSearchMemory64GB: gql.OpenSearchMemoryGb64,
}

var openSearchVersionMap = map[pgratorv1.OpenSearchVersion]gql.OpenSearchMajorVersion{
	pgratorv1.OpenSearchVersionV1:    gql.OpenSearchMajorVersionV1,
	pgratorv1.OpenSearchVersionV2:    gql.OpenSearchMajorVersionV2,
	pgratorv1.OpenSearchVersionV2_19: gql.OpenSearchMajorVersionV219,
	pgratorv1.OpenSearchVersionV3_3:  gql.OpenSearchMajorVersionV33,
}

func openSearchFromCRD(obj *pgratorv1.OpenSearch) (*opensearch.OpenSearch, error) {
	tier, ok := openSearchTierMap[obj.Spec.Tier]
	if !ok {
		return nil, fmt.Errorf("unsupported tier %q", obj.Spec.Tier)
	}

	mem, ok := openSearchMemoryMap[obj.Spec.Memory]
	if !ok {
		return nil, fmt.Errorf("unsupported memory %q", obj.Spec.Memory)
	}

	version, ok := openSearchVersionMap[obj.Spec.Version]
	if !ok {
		return nil, fmt.Errorf("unsupported version %q", obj.Spec.Version)
	}

	return &opensearch.OpenSearch{
		Tier:      tier,
		Memory:    mem,
		Version:   version,
		StorageGB: obj.Spec.StorageGB,
	}, nil
}
