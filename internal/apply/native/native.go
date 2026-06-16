// Package native parses "nais-native" manifests: a stripped-down, user-facing
// flavour of a Kubernetes CRD that hides cluster plumbing such as
// metadata.namespace, metadata.annotations, owner references, generation and
// timestamps. Each parsed resource is either applied through a dedicated
// nais-api mutation (when one exists for its kind) or converted back into a
// native CRD and sent to the generic apply endpoint by the caller.
//
// Regular Kubernetes CRDs (identified by `apiVersion`) are not handled here;
// callers detect them with IsNativeManifest and forward them untouched to the
// apply endpoint.
//
// The native envelope is intentionally minimal:
//
//	version: v1
//	kind: Valkey
//	metadata:
//	  name: my-instance
//	spec:
//	  ...
package native

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// AllowedVersions are the envelope versions the parser understands. Only "v1"
// exists today; "v2" and beyond can be added here as the format evolves.
var allowedVersions = map[string]struct{}{
	"v1": {},
}

// allowedTopLevel and allowedMetadata describe the only fields a stripped
// manifest may contain. Anything else (apiVersion, status, metadata.namespace,
// metadata.annotations, ...) is reported as an ignored field.
var (
	allowedTopLevel = map[string]struct{}{"version": {}, "kind": {}, "metadata": {}, "spec": {}}
	allowedMetadata = map[string]struct{}{"name": {}, "labels": {}}
)

// Metadata identifies where a resource should be applied.
type Metadata struct {
	Name            string
	TeamSlug        string
	EnvironmentName string
	Labels          map[string]string
}

// Action describes what an apply did to a resource.
type Action string

const (
	ActionCreated Action = "created"
	ActionUpdated Action = "updated"
)

// Resource is a single decoded manifest envelope. Spec is kept as a raw YAML
// node so the per-kind handler can decode it into its own typed struct.
type Resource struct {
	Version string
	Kind    string
	Name    string
	Labels  map[string]string
	Spec    yaml.Node

	// IgnoredFields lists envelope fields that are not part of the
	// nais-native format and were dropped (e.g. "metadata.namespace").
	IgnoredFields []string
}

// handler applies a single resource of a given kind through its nais-api
// mutation, using create-or-update (upsert) semantics.
type handler func(ctx context.Context, meta Metadata, spec *yaml.Node) (Action, error)

// registry maps a manifest kind to the mutation-backed handler for it. Kinds
// that are absent here have no dedicated mutation and must be applied through
// the generic CRD apply endpoint instead (see HasMutation).
var registry = map[string]handler{
	"Valkey":     applyValkey,
	"OpenSearch": applyOpenSearch,
}

// HasMutation reports whether the given kind can be applied through a dedicated
// nais-api mutation.
func HasMutation(kind string) bool {
	_, ok := registry[kind]
	return ok
}

// Documents splits raw multi-document YAML into the root mapping node of each
// non-empty document. Callers can inspect each node to classify it (e.g.
// nais-native vs. a regular Kubernetes CRD) before parsing.
func Documents(data []byte) ([]*yaml.Node, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))

	var docs []*yaml.Node
	for {
		var doc yaml.Node
		if err := dec.Decode(&doc); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("failed to decode YAML: %w", err)
		}

		root := documentRoot(&doc)
		if root == nil {
			continue // empty document
		}
		docs = append(docs, root)
	}

	return docs, nil
}

// IsNativeManifest reports whether a decoded document is a nais-native manifest.
// Native manifests use a `version` field, whereas regular Kubernetes CRDs use
// `apiVersion`. A document carrying `apiVersion` is never treated as native.
func IsNativeManifest(root *yaml.Node) bool {
	return !hasKey(root, "apiVersion") && hasKey(root, "version")
}

// ParseDocument parses a single nais-native manifest (the root mapping node of a
// YAML document) into a Resource. It validates the envelope structure (kind,
// metadata.name and a supported version) but does not itself reject ignored
// fields — those are returned in Resource.IgnoredFields so the caller can decide
// whether to error or warn.
func ParseDocument(root *yaml.Node) (Resource, error) {
	var raw struct {
		Version  string `yaml:"version"`
		Kind     string `yaml:"kind"`
		Metadata struct {
			Name   string            `yaml:"name"`
			Labels map[string]string `yaml:"labels,omitempty"`
		} `yaml:"metadata"`
		Spec yaml.Node `yaml:"spec"`
	}
	if err := root.Decode(&raw); err != nil {
		return Resource{}, fmt.Errorf("failed to decode manifest: %w", err)
	}

	if raw.Kind == "" {
		return Resource{}, fmt.Errorf("manifest is missing required field %q", "kind")
	}
	if raw.Metadata.Name == "" {
		return Resource{}, fmt.Errorf("%s manifest is missing required field %q", raw.Kind, "metadata.name")
	}
	if raw.Version == "" {
		return Resource{}, fmt.Errorf("%s/%s is missing required field %q", raw.Kind, raw.Metadata.Name, "version")
	}
	if _, ok := allowedVersions[raw.Version]; !ok {
		return Resource{}, fmt.Errorf("%s/%s has unsupported version %q (supported: %s)", raw.Kind, raw.Metadata.Name, raw.Version, strings.Join(sortedKeys(allowedVersions), ", "))
	}

	return Resource{
		Version:       raw.Version,
		Kind:          raw.Kind,
		Name:          raw.Metadata.Name,
		Spec:          raw.Spec,
		Labels:        raw.Metadata.Labels,
		IgnoredFields: ignoredFields(root),
	}, nil
}

// Parse decodes every nais-native YAML document in data into a Resource.
func Parse(data []byte) ([]Resource, error) {
	docs, err := Documents(data)
	if err != nil {
		return nil, err
	}

	resources := make([]Resource, 0, len(docs))
	for _, root := range docs {
		res, err := ParseDocument(root)
		if err != nil {
			return nil, err
		}
		resources = append(resources, res)
	}

	return resources, nil
}

// Apply applies a single resource through its kind's nais-api mutation using
// create-or-update semantics. It returns an error if the kind has no mutation;
// callers should guard with HasMutation first.
func Apply(ctx context.Context, res Resource, team, environment string) (Action, error) {
	h, ok := registry[res.Kind]
	if !ok {
		return "", fmt.Errorf("kind %q has no nais-api mutation", res.Kind)
	}

	meta := Metadata{
		Name:            res.Name,
		TeamSlug:        team,
		EnvironmentName: environment,
		Labels:          res.Labels,
	}
	spec := res.Spec
	return h(ctx, meta, &spec)
}

// documentRoot returns the mapping node inside a decoded YAML document, or nil
// for an empty document (including an explicit `---` with no body, which decodes
// to a null scalar).
func documentRoot(doc *yaml.Node) *yaml.Node {
	if doc.Kind == yaml.DocumentNode {
		if len(doc.Content) == 0 {
			return nil
		}
		doc = doc.Content[0]
	}
	if doc.Tag == "!!null" || (doc.Kind == yaml.ScalarNode && doc.Value == "") {
		return nil
	}
	return doc
}

// hasKey reports whether a mapping node contains the given top-level key.
func hasKey(root *yaml.Node, key string) bool {
	if root.Kind != yaml.MappingNode {
		return false
	}
	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value == key {
			return true
		}
	}
	return false
}

// ignoredFields walks the envelope mapping and returns the dotted paths of all
// fields that are not part of the nais-native format.
func ignoredFields(root *yaml.Node) []string {
	if root.Kind != yaml.MappingNode {
		return nil
	}

	var ignored []string
	for i := 0; i+1 < len(root.Content); i += 2 {
		key := root.Content[i].Value
		value := root.Content[i+1]

		if _, ok := allowedTopLevel[key]; !ok {
			ignored = append(ignored, key)
			continue
		}

		if key == "metadata" && value.Kind == yaml.MappingNode {
			for j := 0; j+1 < len(value.Content); j += 2 {
				mk := value.Content[j].Value
				if _, ok := allowedMetadata[mk]; !ok {
					ignored = append(ignored, "metadata."+mk)
				}
			}
		}
	}
	return ignored
}

// decodeSpec decodes a spec node into out, rejecting any unknown fields so that
// typos surface as errors instead of being silently dropped.
func decodeSpec(spec *yaml.Node, out any) error {
	if spec == nil || spec.Kind == 0 {
		return fmt.Errorf("manifest is missing required field %q", "spec")
	}

	encoded, err := yaml.Marshal(spec)
	if err != nil {
		return fmt.Errorf("failed to encode spec: %w", err)
	}

	dec := yaml.NewDecoder(bytes.NewReader(encoded))
	dec.KnownFields(true)
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("invalid spec: %w", err)
	}
	return nil
}

// enumValue maps a user-facing CRD value to its GraphQL enum, returning a clear
// error listing the allowed values when the input is unknown.
func enumValue[T ~string](field, raw string, table map[string]T) (T, error) {
	if v, ok := table[raw]; ok {
		return v, nil
	}
	var zero T
	return zero, fmt.Errorf("invalid %s %q (allowed: %s)", field, raw, strings.Join(sortedKeys(table), ", "))
}

// sortedKeys returns the keys of m sorted alphabetically.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
