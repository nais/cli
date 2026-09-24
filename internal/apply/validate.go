package apply

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nais/cli/internal/apply/resource"
	"github.com/nais/cli/internal/validate"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

// ResourceSchemaURL is the aggregate JSON Schema for native Nais manifests.
const ResourceSchemaURL = "https://schemas.nais.io/all.json"

// resourceSchemaLoader allows tests to supply a schema without network access.
var resourceSchemaLoader gojsonschema.JSONLoader

var nativeResourceKinds = map[string]struct{}{
	"Postgres":   {},
	"Valkey":     {},
	"OpenSearch": {},
}

func needsResourceSchemaValidation(root *yaml.Node) bool {
	if root == nil || root.Kind != yaml.MappingNode || !resource.IsNativeManifest(root) {
		return false
	}
	typ, _ := nodeValue(root, "type")
	_, ok := nativeResourceKinds[typ]
	return ok
}

// nodeValue returns the scalar value of a top-level key in a mapping node.
func nodeValue(root *yaml.Node, key string) (string, bool) {
	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value == key {
			return root.Content[i+1].Value, true
		}
	}
	return "", false
}

// ValidateNativeManifests checks supported native manifests before any apply
// mutations. A schema download error is returned separately so callers can
// warn and proceed; invalid manifests return an error and must not be applied.
func ValidateNativeManifests(ctx context.Context, docs []*yaml.Node, allowIgnoredFields bool) (warning, validationErr error) {
	var schema *gojsonschema.Schema
	var errs []string
	for _, doc := range docs {
		if !needsResourceSchemaValidation(doc) {
			continue
		}
		manifest, parseErr := resource.ParseManifest(doc)
		if parseErr == nil && len(manifest.IgnoredFields) > 0 {
			if !allowIgnoredFields {
				errs = append(errs, handleIgnoredFields(manifest, false, nil).Error())
				continue
			}
			doc = withoutIgnoredFields(doc, manifest.IgnoredFields)
		}
		if schema == nil && warning == nil {
			schema, warning = loadResourceSchema(ctx)
		}
		if warning != nil {
			continue
		}
		schemaErrs, err := validateResourceSchema(doc, schema)
		if err != nil {
			return warning, fmt.Errorf("%s: %w", docLabel(doc), err)
		}
		if len(schemaErrs) > 0 {
			errs = append(errs, fmt.Sprintf("%s:\n%s", docLabel(doc), formatResourceSchemaErrors(schemaErrs)))
		}
	}
	if len(errs) > 0 {
		return warning, fmt.Errorf("schema validation failed for %d manifest(s):\n  %s", len(errs), strings.Join(errs, "\n  "))
	}
	return warning, nil
}

func withoutIgnoredFields(doc *yaml.Node, ignoredFields []string) *yaml.Node {
	ignored := make(map[string]bool, len(ignoredFields))
	for _, field := range ignoredFields {
		ignored[field] = true
	}
	filtered := *doc
	filtered.Content = make([]*yaml.Node, 0, len(doc.Content))
	for i := 0; i+1 < len(doc.Content); i += 2 {
		if !ignored[doc.Content[i].Value] {
			filtered.Content = append(filtered.Content, doc.Content[i], doc.Content[i+1])
		}
	}
	return &filtered
}

func docLabel(doc *yaml.Node) string {
	typ, _ := nodeValue(doc, "type")
	name, _ := nodeValue(doc, "name")
	return fmt.Sprintf("%s/%s", typ, name)
}

func loadResourceSchema(ctx context.Context) (*gojsonschema.Schema, error) {
	if resourceSchemaLoader != nil {
		return gojsonschema.NewSchema(resourceSchemaLoader)
	}
	return loadPublishedSchema(ctx, ResourceSchemaURL)
}

func loadPublishedSchema(ctx context.Context, schemaURL string) (*gojsonschema.Schema, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	client := &http.Client{}
	fetch := func(address string) ([]byte, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
		if err != nil {
			return nil, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("fetch %s: %s", address, resp.Status)
		}
		return io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	}

	base, err := url.Parse(schemaURL)
	if err != nil {
		return nil, err
	}
	root, err := fetch(schemaURL)
	if err != nil {
		return nil, err
	}
	var aggregate struct {
		OneOf []struct {
			Ref string `json:"$ref"`
		} `json:"oneOf"`
	}
	if err := json.Unmarshal(root, &aggregate); err != nil {
		return nil, err
	}
	if len(aggregate.OneOf) == 0 {
		return nil, fmt.Errorf("resource schema has no oneOf entries")
	}
	loader := gojsonschema.NewSchemaLoader()
	for _, entry := range aggregate.OneOf {
		ref, err := url.Parse(entry.Ref)
		if err != nil || entry.Ref == "" || ref.IsAbs() || ref.Host != "" {
			return nil, fmt.Errorf("invalid resource schema reference %q", entry.Ref)
		}
		address := base.ResolveReference(ref).String()
		data, err := fetch(address)
		if err != nil {
			return nil, err
		}
		if err := loader.AddSchema(address, gojsonschema.NewBytesLoader(data)); err != nil {
			return nil, err
		}
	}
	if err := loader.AddSchema(schemaURL, gojsonschema.NewBytesLoader(root)); err != nil {
		return nil, err
	}
	return loader.Compile(gojsonschema.NewReferenceLoader(schemaURL))
}

// validateResourceSchema validates a routed document against the loaded schema.
func validateResourceSchema(doc *yaml.Node, schema *gojsonschema.Schema) ([]gojsonschema.ResultError, error) {
	encoded, err := yaml.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to encode document: %w", err)
	}

	messages, err := validate.YAMLToJSONMessages(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to convert document to JSON: %w", err)
	}
	if len(messages) == 0 {
		return nil, nil
	}

	documentLoader := gojsonschema.NewBytesLoader(messages[0])
	result, err := schema.Validate(documentLoader)
	if err != nil {
		return nil, err
	}

	if result.Valid() {
		return nil, nil
	}
	return result.Errors(), nil
}

// formatResourceSchemaErrors renders schema validation errors the same way
// `nais validate` does: one indented "field: description" pair per error, with
// the noisy oneOf root error skipped.
func formatResourceSchemaErrors(errs []gojsonschema.ResultError) string {
	var lines []string
	for _, err := range errs {
		if err.Field() == gojsonschema.STRING_ROOT_SCHEMA_PROPERTY && err.Type() == "number_one_of" {
			continue
		}
		lines = append(lines, fmt.Sprintf("  %q: %s", err.Field(), err.Description()))
	}
	return strings.Join(lines, "\n")
}
