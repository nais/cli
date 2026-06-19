package resource

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

// allowedVersions are the envelope versions the parser understands. "v2" and
// beyond can be added here as the format evolves.
var allowedVersions = map[string]struct{}{
	"v1": {},
}

// allowedTopLevel and allowedMetadata are the only fields a stripped manifest may
// contain; anything else is reported as an ignored field.
var (
	allowedTopLevel = map[string]struct{}{"version": {}, "kind": {}, "metadata": {}, "spec": {}, "data": {}, "binaryData": {}}
	allowedMetadata = map[string]struct{}{"name": {}, "labels": {}}
)

// Manifest is a decoded nais-native manifest envelope. Spec is kept as a raw YAML
// node so each resource can decode it into its own typed struct.
type Manifest struct {
	Version string
	Kind    string
	Name    string
	Labels  map[string]string
	Spec    yaml.Node

	// Data holds top-level key-value data for resources that use data/binaryData
	// instead of spec (e.g. Config).
	Data       map[string]string
	BinaryData map[string]string

	// IgnoredFields are envelope fields that are not part of the nais-native
	// format (e.g. "metadata.namespace").
	IgnoredFields []string
}

// Documents splits raw multi-document YAML into the root mapping node of each
// non-empty document, for the caller to classify and parse.
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

// IsNativeManifest reports whether a document is a nais-native manifest, which
// uses `version` rather than the `apiVersion` of a regular Kubernetes CRD.
func IsNativeManifest(root *yaml.Node) bool {
	return !hasKey(root, "apiVersion") && hasKey(root, "version")
}

// ParseManifest parses a nais-native manifest into a Manifest, validating the
// envelope (kind, metadata.name, supported version). Fields outside the format
// are not rejected here but returned in Manifest.IgnoredFields for the caller to
// error or warn on.
func ParseManifest(root *yaml.Node) (Manifest, error) {
	var raw struct {
		Version  string `yaml:"version"`
		Kind     string `yaml:"kind"`
		Metadata struct {
			Name   string            `yaml:"name"`
			Labels map[string]string `yaml:"labels,omitempty"`
		} `yaml:"metadata"`
		Spec       yaml.Node         `yaml:"spec"`
		Data       map[string]string `yaml:"data,omitempty"`
		BinaryData map[string]string `yaml:"binaryData,omitempty"`
	}
	if err := root.Decode(&raw); err != nil {
		return Manifest{}, fmt.Errorf("failed to decode manifest: %w", err)
	}

	if raw.Kind == "" {
		return Manifest{}, fmt.Errorf("manifest is missing required field %q", "kind")
	}
	if raw.Metadata.Name == "" {
		return Manifest{}, fmt.Errorf("%s manifest is missing required field %q", raw.Kind, "metadata.name")
	}
	if raw.Version == "" {
		return Manifest{}, fmt.Errorf("%s/%s is missing required field %q", raw.Kind, raw.Metadata.Name, "version")
	}
	if _, ok := allowedVersions[raw.Version]; !ok {
		return Manifest{}, fmt.Errorf("%s/%s has unsupported version %q (supported: %s)", raw.Kind, raw.Metadata.Name, raw.Version, strings.Join(sortedKeys(allowedVersions), ", "))
	}

	return Manifest{
		Version:       raw.Version,
		Kind:          raw.Kind,
		Name:          raw.Metadata.Name,
		Spec:          raw.Spec,
		Labels:        raw.Metadata.Labels,
		Data:          raw.Data,
		BinaryData:    raw.BinaryData,
		IgnoredFields: ignoredFields(root),
	}, nil
}

// Parse decodes every nais-native YAML document in data.
func Parse(data []byte) ([]Manifest, error) {
	docs, err := Documents(data)
	if err != nil {
		return nil, err
	}

	manifests := make([]Manifest, 0, len(docs))
	for _, root := range docs {
		m, err := ParseManifest(root)
		if err != nil {
			return nil, err
		}
		manifests = append(manifests, m)
	}

	return manifests, nil
}

// documentRoot returns the mapping node inside a decoded YAML document, or nil
// for an empty document (e.g. a bare `---`).
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

// hasKey reports whether a mapping node has the given top-level key.
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

// ignoredFields returns the dotted paths of all envelope fields that are not
// part of the nais-native format.
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

// decodeSpec decodes a spec node into out, rejecting unknown fields so typos
// surface as errors instead of being silently dropped.
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
