package apply

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nais/cli/internal/apply/command/flag"
	"github.com/nais/cli/internal/apply/native"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// crdGroup is the Kubernetes API group used when converting a nais-native
// manifest back into a CRD for the generic apply endpoint.
const crdGroup = "nais.io"

func Run(ctx context.Context, filePath string, flags *flag.Apply, out *naistrix.OutputWriter) error {
	environment, err := resolveEnvironment(ctx, string(flags.Environment), out)
	if err != nil {
		return err
	}

	data, err := render(filePath, string(flags.Mixin), environment, flags.Set, out)
	if err != nil {
		return err
	}

	docs, err := native.Documents(data)
	if err != nil {
		return err
	}
	if len(docs) == 0 {
		return fmt.Errorf("no resources found in %s", filePath)
	}

	var (
		crds []unstructured.Unstructured
		errs []string
	)

	for _, doc := range docs {
		// Regular Kubernetes CRDs (identified by apiVersion) are always
		// forwarded to the generic apply endpoint untouched, and never sent
		// to a mutation.
		if !native.IsNativeManifest(doc) {
			crd, err := decodeCRD(doc)
			if err != nil {
				errs = append(errs, err.Error())
				continue
			}
			crds = append(crds, crd)
			continue
		}

		res, err := native.ParseDocument(doc)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}

		if err := handleIgnoredFields(res, flags.AllowIgnoredFields, out); err != nil {
			return err
		}

		// nais-native kinds without a dedicated mutation are converted back
		// into a CRD and applied through the generic endpoint.
		if !native.HasMutation(res.Kind) {
			crd, err := toUnstructured(res)
			if err != nil {
				errs = append(errs, fmt.Sprintf("%s/%s: %v", res.Kind, res.Name, err))
				continue
			}
			crds = append(crds, crd)
			continue
		}

		action, err := native.Apply(ctx, res, flags.Team, environment)
		if err != nil {
			out.Warnf("%s/%s: %v\n", res.Kind, res.Name, err)
			errs = append(errs, fmt.Sprintf("%s/%s: %v", res.Kind, res.Name, err))
			continue
		}
		out.Successf("%s/%s: %s\n", res.Kind, res.Name, action)
	}

	if len(crds) > 0 {
		errs = append(errs, applyCRDs(ctx, flags.Team, environment, crds, out)...)
	}

	if len(errs) > 0 {
		return fmt.Errorf("apply failed for %d resource(s):\n  %s", len(errs), strings.Join(errs, "\n  "))
	}

	return nil
}

// handleIgnoredFields reports nais-native manifest fields that nais apply does
// not use. By default this is a hard error; with --allow-ignored-fields it is
// downgraded to a warning.
func handleIgnoredFields(r native.Resource, allow bool, out *naistrix.OutputWriter) error {
	if len(r.IgnoredFields) == 0 {
		return nil
	}
	fields := strings.Join(r.IgnoredFields, ", ")
	if allow {
		out.Warnf("%s/%s: ignoring fields not used by nais apply: %s\n", r.Kind, r.Name, fields)
		return nil
	}
	return fmt.Errorf(
		"%s/%s contains fields not used by nais apply: %s\nRemove them, or pass --allow-ignored-fields to ignore them with a warning instead",
		r.Kind, r.Name, fields,
	)
}

// applyCRDs sends CRDs to the generic apply endpoint and returns any
// per-resource errors.
func applyCRDs(ctx context.Context, team, environment string, crds []unstructured.Unstructured, out *naistrix.OutputWriter) []string {
	resp, err := naisapi.ApplyManifests(ctx, team, environment, crds)
	if err != nil {
		return []string{fmt.Sprintf("failed to apply manifests: %v", err)}
	}

	var errs []string
	for _, r := range resp.Results {
		if r.Status == "error" {
			out.Warnf("%s: %s\n", r.Resource, r.Error)
			errs = append(errs, fmt.Sprintf("%s: %s", r.Resource, r.Error))
			continue
		}
		out.Successf("%s: %s\n", r.Resource, r.Status)
	}
	return errs
}

// decodeCRD decodes a regular Kubernetes CRD document into an unstructured
// object, preserving all of its fields (apiVersion, metadata, labels, etc.).
func decodeCRD(doc *yaml.Node) (unstructured.Unstructured, error) {
	var u unstructured.Unstructured
	if err := doc.Decode(&u.Object); err != nil {
		return unstructured.Unstructured{}, fmt.Errorf("failed to decode manifest: %w", err)
	}
	if u.GetKind() == "" {
		return unstructured.Unstructured{}, fmt.Errorf("manifest is missing required field %q", "kind")
	}
	if u.GetName() == "" {
		return unstructured.Unstructured{}, fmt.Errorf("%s manifest is missing required field %q", u.GetKind(), "metadata.name")
	}
	return u, nil
}

// toUnstructured rebuilds a native Kubernetes CRD from a stripped manifest so it
// can be sent to the generic apply endpoint.
func toUnstructured(r native.Resource) (unstructured.Unstructured, error) {
	spec := map[string]any{}
	if r.Spec.Kind != 0 {
		if err := r.Spec.Decode(&spec); err != nil {
			return unstructured.Unstructured{}, fmt.Errorf("failed to decode spec: %w", err)
		}
	}

	return unstructured.Unstructured{Object: map[string]any{
		"apiVersion": crdGroup + "/" + r.Version,
		"kind":       r.Kind,
		"metadata":   map[string]any{"name": r.Name},
		"spec":       spec,
	}}, nil
}

// readManifestFile reads a YAML manifest file, validating the extension.
func readManifestFile(filePath string) ([]byte, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	switch strings.TrimLeft(filepath.Ext(filePath), ".") {
	case "yaml", "yml":
	default:
		return nil, fmt.Errorf("unsupported file extension for file %s (expected .yaml or .yml)", filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return data, nil
}
