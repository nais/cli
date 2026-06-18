package apply

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nais/cli/internal/apply/command/flag"
	"github.com/nais/cli/internal/apply/resource"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// waitTarget is a resource to poll for readiness after a successful apply.
type waitTarget struct {
	waiter resource.Waiter
	name   string
}

// crdGroup is the fallback API group for stripped manifests of unknown kinds.
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

	docs, err := resource.Documents(data)
	if err != nil {
		return err
	}
	if len(docs) == 0 {
		return fmt.Errorf("no resources found in %s", filePath)
	}

	var (
		crds        []unstructured.Unstructured
		errs        []string
		waitTargets []waitTarget
	)

	// Captured before applying so a rollout from this apply can be told apart
	// from the resource's previous state while waiting.
	since := time.Now()

	for _, doc := range docs {
		// Regular CRDs are forwarded to the generic apply endpoint untouched.
		if !resource.IsNativeManifest(doc) {
			crd, err := decodeCRD(doc)
			if err != nil {
				errs = append(errs, err.Error())
				continue
			}
			if flags.DryRun {
				out.Printf("%s/%s: would apply\n", crd.GetKind(), crd.GetName())
				printDryRunYAML(doc, out)
				continue
			}
			crds = append(crds, crd)
			r, _ := resource.ForCRD(crd.GetAPIVersion(), crd.GetKind())
			waitTargets = appendWaitTarget(waitTargets, r, crd.GetName())
			continue
		}

		m, err := resource.ParseManifest(doc)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}

		if err := handleIgnoredFields(m, flags.AllowIgnoredFields, out); err != nil {
			return err
		}

		r, _ := resource.ForManifest(m)
		if flags.DryRun {
			// Keep conversion/validation behavior aligned with real apply for
			// non-mutation resources.
			if _, ok := r.(resource.Applier); !ok {
				if _, err := toUnstructured(m, r); err != nil {
					errs = append(errs, fmt.Sprintf("%s/%s: %v", m.Kind, m.Name, err))
					continue
				}
			}
			out.Printf("%s/%s: would apply\n", m.Kind, m.Name)
			printDryRunYAML(doc, out)
			continue
		}

		if applier, ok := r.(resource.Applier); ok {
			action, err := applier.Apply(ctx, resource.Metadata{
				Name:            m.Name,
				TeamSlug:        flags.Team,
				EnvironmentName: environment,
				Labels:          m.Labels,
			}, &m.Spec)
			if err != nil {
				out.Warnf("%s/%s: %v\n", m.Kind, m.Name, err)
				errs = append(errs, fmt.Sprintf("%s/%s: %v", m.Kind, m.Name, err))
				continue
			}
			waitTargets = appendWaitTarget(waitTargets, r, m.Name)
			out.Successf("%s/%s: %s\n", m.Kind, m.Name, action)
			continue
		}

		// No mutation: convert back into a CRD for the generic endpoint.
		crd, err := toUnstructured(m, r)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s/%s: %v", m.Kind, m.Name, err))
			continue
		}
		crds = append(crds, crd)
		waitTargets = appendWaitTarget(waitTargets, r, m.Name)
	}

	if len(crds) > 0 {
		errs = append(errs, applyCRDs(ctx, flags.Team, environment, crds, out)...)
	}

	if len(errs) > 0 {
		return fmt.Errorf("apply failed for %d resource(s):\n  %s", len(errs), strings.Join(errs, "\n  "))
	}

	if flags.DryRun {
		out.Printf("dry-run complete: no resources were applied\n")
		return nil
	}

	if flags.Wait {
		if err := waitForReady(ctx, flags.Team, environment, waitTargets, since, flags.Timeout, out); err != nil {
			return err
		}
	}

	return nil
}

// appendWaitTarget records a resource to wait on, skipping a nil resource (e.g.
// an unexpected apiVersion) or one without a Waiter (e.g. Valkey, OpenSearch).
func appendWaitTarget(targets []waitTarget, r resource.Resource, name string) []waitTarget {
	waiter, ok := r.(resource.Waiter)
	if !ok {
		return targets
	}
	return append(targets, waitTarget{waiter: waiter, name: name})
}

// waitForReady polls every wait target until it is ready, sharing one timeout.
func waitForReady(ctx context.Context, team, environment string, targets []waitTarget, since time.Time, timeout time.Duration, out *naistrix.OutputWriter) error {
	if len(targets) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var errs []string
	for _, t := range targets {
		if err := t.waiter.Wait(ctx, team, environment, t.name, since, out); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("wait failed for %d resource(s):\n  %s", len(errs), strings.Join(errs, "\n  "))
	}
	return nil
}

// handleIgnoredFields errors on manifest fields nais apply does not use, or warns
// instead when --allow-ignored-fields is set.
func handleIgnoredFields(m resource.Manifest, allow bool, out *naistrix.OutputWriter) error {
	if len(m.IgnoredFields) == 0 {
		return nil
	}
	fields := strings.Join(m.IgnoredFields, ", ")
	if allow {
		out.Warnf("%s/%s: ignoring fields not used by nais apply: %s\n", m.Kind, m.Name, fields)
		return nil
	}
	return fmt.Errorf(
		"%s/%s contains fields not used by nais apply: %s\nRemove them, or pass --allow-ignored-fields to ignore them with a warning instead",
		m.Kind, m.Name, fields,
	)
}

// applyCRDs sends CRDs to the generic apply endpoint, returning per-resource
// errors.
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

// decodeCRD decodes a regular CRD document into an unstructured object,
// preserving all of its fields.
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

// toUnstructured builds a CRD from a stripped manifest for the generic apply
// endpoint. The apiVersion comes from the resolved resource, falling back to the
// nais.io group plus the manifest version for unknown kinds.
func toUnstructured(m resource.Manifest, r resource.Resource) (unstructured.Unstructured, error) {
	spec := map[string]any{}
	if m.Spec.Kind != 0 {
		if err := m.Spec.Decode(&spec); err != nil {
			return unstructured.Unstructured{}, fmt.Errorf("failed to decode spec: %w", err)
		}
	}

	apiVersion := crdGroup + "/" + m.Version
	if r != nil && r.APIVersion() != "" {
		apiVersion = r.APIVersion()
	}

	return unstructured.Unstructured{Object: map[string]any{
		"apiVersion": apiVersion,
		"kind":       m.Kind,
		"metadata":   map[string]any{"name": m.Name},
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

func printDryRunYAML(doc *yaml.Node, out *naistrix.OutputWriter) {
	rendered, err := yaml.Marshal(doc)
	if err != nil {
		out.Printf("failed to render dry-run YAML: %v\n", err)
		return
	}

	out.Printf("---\n%s", rendered)
	if len(rendered) == 0 || rendered[len(rendered)-1] != '\n' {
		out.Printf("\n")
	}
}
