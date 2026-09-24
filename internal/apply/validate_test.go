package apply

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	applyflag "github.com/nais/cli/internal/apply/command/flag"
	flagspkg "github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
	"github.com/xeipuuv/gojsonschema"
)

// testResourceSchema models the native manifest envelopes in all.json.
const testResourceSchema = `{
  "oneOf": [
    {
      "type": "object",
      "additionalProperties": false,
      "required": ["version", "type", "name", "spec"],
      "properties": {
        "version": {"const": "v1"},
        "type": {"const": "Valkey"},
        "name": {"type": "string"},
        "labels": {"type": "object"},
        "spec": {
          "type": "object",
          "additionalProperties": false,
          "required": ["memory", "tier"],
          "properties": {
            "memory": {"enum": ["1GB", "4GB"]},
            "tier": {"enum": ["SingleNode", "HighAvailability"]}
          }
        }
      }
    },
    {
      "type": "object",
      "additionalProperties": false,
      "required": ["version", "type", "name", "spec"],
      "properties": {
        "version": {"const": "v1"},
        "type": {"const": "Postgres"},
        "name": {"type": "string"},
        "spec": {
          "type": "object",
          "additionalProperties": false,
          "required": ["majorVersion"],
          "properties": {
            "majorVersion": {"enum": ["16", "17", "18"]}
          }
        }
      }
    },
    {
      "type": "object",
      "additionalProperties": false,
      "required": ["version", "type", "name", "spec"],
      "properties": {
        "version": {"const": "v1"},
        "type": {"const": "OpenSearch"},
        "name": {"type": "string"},
        "spec": {"type": "object"}
      }
    }
  ]
}`

// withTestResourceSchema swaps the resource schema loader for the test fixture
// (or an arbitrary override) and restores the original after the test.
func withTestResourceSchema(t *testing.T, loader gojsonschema.JSONLoader) {
	t.Helper()
	original := resourceSchemaLoader
	resourceSchemaLoader = loader
	t.Cleanup(func() {
		resourceSchemaLoader = original
	})
}

func writeManifestAndDryRun(t *testing.T, manifest string) (string, error) {
	t.Helper()
	return writeManifestAndRun(t, manifest, true, false)
}

func writeManifestAndRun(t *testing.T, manifest string, dryRun, allowIgnoredFields bool) (string, error) {
	t.Helper()
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "nais.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	var out bytes.Buffer
	flags := &applyflag.Apply{
		GlobalFlags: &flagspkg.GlobalFlags{
			AdditionalFlags: &flagspkg.AdditionalFlags{
				Team:        "my-team",
				Environment: "dev",
			},
		},
		DryRun:             dryRun,
		AllowIgnoredFields: allowIgnoredFields,
	}

	err := Run(context.Background(), manifestPath, flags, naistrix.NewOutputWriter(&out, new(naistrix.Count)))
	return out.String(), err
}

func TestRun_DryRunValidatesNativeValkey(t *testing.T) {
	withTestResourceSchema(t, gojsonschema.NewStringLoader(testResourceSchema))

	out, err := writeManifestAndDryRun(t, `
version: v1
type: Valkey
name: my-valkey
spec:
  memory: 1GB
  tier: SingleNode
`)
	if err != nil {
		t.Fatalf("Run: %v\noutput:\n%s", err, out)
	}
	if !strings.Contains(out, "Valkey/my-valkey: would apply") {
		t.Errorf("output %q does not contain expected apply line", out)
	}
}

func TestRun_DryRunRejectsInvalidNativeValkeyEnum(t *testing.T) {
	withTestResourceSchema(t, gojsonschema.NewStringLoader(testResourceSchema))

	_, err := writeManifestAndDryRun(t, `
version: v1
type: Valkey
name: my-valkey
spec:
  memory: 999GB
  tier: SingleNode
`)
	mustErrorContains(t, err, "Valkey/my-valkey")
}

func TestRun_DryRunRejectsUnknownNativeValkeySpecField(t *testing.T) {
	withTestResourceSchema(t, gojsonschema.NewStringLoader(testResourceSchema))

	_, err := writeManifestAndDryRun(t, `
version: v1
type: Valkey
name: my-valkey
spec:
  memory: 1GB
  tier: SingleNode
  bogusField: nope
`)
	mustErrorContains(t, err, "Valkey/my-valkey")
}

func TestRun_DryRunRejectsUnknownTopLevelField(t *testing.T) {
	withTestResourceSchema(t, gojsonschema.NewStringLoader(testResourceSchema))

	_, err := writeManifestAndDryRun(t, `
version: v1
type: Valkey
name: my-valkey
bogusTopLevel: nope
spec:
  memory: 1GB
  tier: SingleNode
`)
	mustErrorContains(t, err, "Valkey/my-valkey")
	mustErrorContains(t, err, "bogusTopLevel")
	mustErrorContains(t, err, "--allow-ignored-fields")
}

func TestRun_DryRunValidatesNativePostgres(t *testing.T) {
	withTestResourceSchema(t, gojsonschema.NewStringLoader(testResourceSchema))

	out, err := writeManifestAndDryRun(t, `
version: v1
type: Postgres
name: my-postgres
spec:
  majorVersion: "18"
`)
	if err != nil {
		t.Fatalf("Run: %v\noutput:\n%s", err, out)
	}
	if !strings.Contains(out, "Postgres/my-postgres: would apply") {
		t.Errorf("output %q does not contain expected apply line", out)
	}
}

func TestRun_DryRunRejectsInvalidNativePostgres(t *testing.T) {
	withTestResourceSchema(t, gojsonschema.NewStringLoader(testResourceSchema))

	_, err := writeManifestAndDryRun(t, `
version: v1
type: Postgres
name: my-postgres
spec:
  majorVersion: "99"
`)
	mustErrorContains(t, err, "Postgres/my-postgres")
}

func TestRun_DryRunSkipsCRDPostgres(t *testing.T) {
	withTestResourceSchema(t, gojsonschema.NewStringLoader(testResourceSchema))

	out, err := writeManifestAndDryRun(t, `
apiVersion: nais.io/v1
kind: Postgres
metadata:
  name: my-postgres
spec:
  majorVersion: "18"
`)
	if err != nil {
		t.Fatalf("Run: %v\noutput:\n%s", err, out)
	}
	if !strings.Contains(out, "Postgres/my-postgres: would apply") {
		t.Errorf("output %q does not contain expected apply line", out)
	}
}

func TestRun_DryRunRejectsInvalidNativeOpenSearch(t *testing.T) {
	withTestResourceSchema(t, gojsonschema.NewStringLoader(testResourceSchema))

	_, err := writeManifestAndDryRun(t, `
version: v1
type: OpenSearch
name: my-search
unknownField: nope
spec: {}
`)
	mustErrorContains(t, err, "OpenSearch/my-search")
}

func TestRun_ValidatesBeforeApplyingAnyManifest(t *testing.T) {
	withTestResourceSchema(t, gojsonschema.NewStringLoader(testResourceSchema))

	for _, dryRun := range []bool{false, true} {
		name := "apply"
		if dryRun {
			name = "dry-run"
		}
		t.Run(name, func(t *testing.T) {
			out, err := writeManifestAndRun(t, `
version: v1
type: Valkey
name: first
spec:
  memory: 1GB
  tier: SingleNode
---
version: v1
type: Postgres
name: invalid
spec:
  majorVersion: "99"
`, dryRun, false)
			mustErrorContains(t, err, "Postgres/invalid")
			if strings.Contains(out, "Valkey/first: would apply") || strings.Contains(out, "Valkey/first: created") {
				t.Errorf("first resource was processed before validation failed: %s", out)
			}
		})
	}
}

func TestRun_RejectsIgnoredFieldsBeforeApplying(t *testing.T) {
	withTestResourceSchema(t, gojsonschema.NewStringLoader(testResourceSchema))

	for _, dryRun := range []bool{false, true} {
		out, err := writeManifestAndRun(t, `
version: v1
type: Valkey
name: first
spec:
  memory: 1GB
  tier: SingleNode
---
version: v1
type: Valkey
name: second
namespace: ignored
spec:
  memory: 1GB
  tier: SingleNode
`, dryRun, false)
		mustErrorContains(t, err, "Valkey/second contains fields not used by nais apply: namespace")
		mustErrorContains(t, err, "--allow-ignored-fields")
		if strings.Contains(out, "Valkey/first: would apply") || strings.Contains(out, "Valkey/first: created") {
			t.Errorf("first resource was processed before validation failed: %s", out)
		}
	}
}

func TestRun_RejectsIgnoredValkeyPersistence(t *testing.T) {
	withTestResourceSchema(t, gojsonschema.NewStringLoader(testResourceSchema))

	for _, dryRun := range []bool{false, true} {
		_, err := writeManifestAndRun(t, `
version: v1
type: Valkey
name: my-valkey
spec:
  memory: 1GB
  tier: SingleNode
  persistence:
    disabled: true
`, dryRun, false)
		mustErrorContains(t, err, "persistence")
	}
}

func TestRun_AllowIgnoredFieldsPreservesNativeManifestBehavior(t *testing.T) {
	withTestResourceSchema(t, gojsonschema.NewStringLoader(testResourceSchema))

	out, err := writeManifestAndRun(t, `
version: v1
type: Valkey
name: my-valkey
namespace: ignored
spec:
  memory: 1GB
  tier: SingleNode
`, true, true)
	if err != nil {
		t.Fatalf("Run: %v\noutput:\n%s", err, out)
	}
	if !strings.Contains(out, "ignoring fields not used by nais apply: namespace") {
		t.Errorf("expected warning for ignored field, got %q", out)
	}
	if !strings.Contains(out, "Valkey/my-valkey: would apply") {
		t.Errorf("expected dry-run result, got %q", out)
	}
}

func TestLoadPublishedSchemaResolvesReferences(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		switch r.URL.Path {
		case "/all.json":
			_, _ = w.Write([]byte(`{"oneOf":[{"$ref":"postgres.json"}]}`))
		case "/postgres.json":
			_, _ = w.Write([]byte(`{"type":"object","required":["type"],"properties":{"type":{"const":"Postgres"}}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	schema, err := loadPublishedSchema(context.Background(), server.URL+"/all.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		doc   string
		valid bool
	}{
		{`{"type":"Postgres"}`, true},
		{`{"type":"Other"}`, false},
	} {
		result, err := schema.Validate(gojsonschema.NewStringLoader(tc.doc))
		if err != nil {
			t.Fatal(err)
		}
		if result.Valid() != tc.valid {
			t.Errorf("validating %s: valid=%v, want %v", tc.doc, result.Valid(), tc.valid)
		}
	}
	if got := requests.Load(); got != 2 {
		t.Errorf("fetched %d times, want once per schema file", got)
	}
}

func TestLoadPublishedSchemaRejectsEmptyAggregate(t *testing.T) {
	for _, body := range []string{`{}`, `{"oneOf":[]}`} {
		t.Run(body, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(body))
			}))
			defer server.Close()

			_, err := loadPublishedSchema(context.Background(), server.URL+"/all.json")
			mustErrorContains(t, err, "resource schema has no oneOf entries")
		})
	}
}

func TestRun_DryRunSkipsUnroutedApplicationCRD(t *testing.T) {
	withTestResourceSchema(t, gojsonschema.NewStringLoader(testResourceSchema))

	out, err := writeManifestAndDryRun(t, `
apiVersion: nais.io/v1alpha1
kind: Application
metadata:
  name: myapp
spec:
  image: ghcr.io/nais/app:latest
  thisFieldDoesNotExistInAnySchema: true
`)
	if err != nil {
		t.Fatalf("Run: %v\noutput:\n%s", err, out)
	}
	if !strings.Contains(out, "Application/myapp: would apply") {
		t.Errorf("output %q does not contain expected apply line", out)
	}
}

func TestRun_DryRunSkipsUnroutedNativeConfig(t *testing.T) {
	withTestResourceSchema(t, gojsonschema.NewStringLoader(testResourceSchema))

	out, err := writeManifestAndDryRun(t, `
version: v1
type: Config
name: my-config
data:
  someKey: someValue
`)
	if err != nil {
		t.Fatalf("Run: %v\noutput:\n%s", err, out)
	}
	if !strings.Contains(out, "Config/my-config: would apply") {
		t.Errorf("output %q does not contain expected apply line", out)
	}
}

func TestRun_DryRunToleratesUnavailableSchema(t *testing.T) {
	// A reference loader pointing at a nonexistent file fails when loading,
	// simulating an unreachable/unpublished schema (e.g. all.json 404, offline).
	withTestResourceSchema(t, gojsonschema.NewReferenceLoader("file:///nonexistent/schema.json"))

	out, err := writeManifestAndDryRun(t, `
version: v1
type: Valkey
name: my-valkey
spec:
  memory: not-a-valid-value-at-all
  tier: not-a-valid-tier-either
`)
	if err != nil {
		t.Fatalf("Run: %v\noutput:\n%s", err, out)
	}
	if !strings.Contains(out, "schema validation unavailable") {
		t.Errorf("output %q does not contain schema-unavailable warning", out)
	}
	if !strings.Contains(out, "Valkey/my-valkey: would apply") {
		t.Errorf("output %q does not contain expected apply line", out)
	}
}
