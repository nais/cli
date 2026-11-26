package validate

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nais/naistrix"
	"github.com/xeipuuv/gojsonschema"
)

//go:embed schema.json
var schema []byte

func TestValidateSuggestsDryRunForNativeResources(t *testing.T) {
	for _, tc := range []struct {
		name   string
		prefix string
		kind   string
	}{
		{name: "Postgres", kind: "Postgres"},
		{name: "Valkey", kind: "Valkey"},
		{name: "OpenSearch", kind: "OpenSearch"},
		{name: "mixed documents", prefix: "apiVersion: nais.io/v1alpha1\nkind: Application\n---\n", kind: "Valkey"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "manifest.yaml")
			manifest := fmt.Sprintf("%sversion: v1\ntype: %s\nname: example\n", tc.prefix, tc.kind)
			if err := os.WriteFile(path, []byte(manifest), 0o600); err != nil {
				t.Fatalf("write manifest: %v", err)
			}

			v := New([]string{path})
			v.SchemaLoader = gojsonschema.NewBytesLoader(schema)
			level := naistrix.OutputVerbosityLevelNormal
			err := v.Validate(naistrix.NewOutputWriter(io.Discard, io.Discard, &level))
			if want := tc.kind + " manifests cannot be validated with nais validate"; err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("validation error = %v, want it to contain %q", err, want)
			}
			if want := "use nais apply --dry-run instead"; !strings.Contains(err.Error(), want) {
				t.Errorf("validation error = %v, want it to contain %q", err, want)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	schemaLoader := gojsonschema.NewBytesLoader(schema)

	jsonVars, err := TemplateVariablesFromFile("testdata/vars.json")
	if err != nil {
		t.Fatalf("load JSON variables: %v", err)
	}
	if len(jsonVars) == 0 {
		t.Fatal("expected JSON variables")
	}

	yamlVars, err := TemplateVariablesFromFile("testdata/vars.yaml")
	if err != nil {
		t.Fatalf("load YAML variables: %v", err)
	}
	if len(yamlVars) == 0 {
		t.Fatal("expected YAML variables")
	}

	sliceVars := TemplateVariablesFromSlice([]string{
		"app=some-app",
		"namespace=some-namespace",
		"image=some-image",
		"team=some-team",
	})
	for _, key := range []string{"app", "namespace", "image", "team"} {
		if _, ok := sliceVars[key]; !ok {
			t.Fatalf("missing template variable %q", key)
		}
	}

	for name, test := range map[string]struct {
		path    string
		vars    TemplateVariables
		wantErr bool
	}{
		"valid": {
			path: "testdata/nais-valid.yaml",
		},
		"valid multi-document": {
			path: "testdata/nais-valid-multidocument.yaml",
		},
		"valid template with json vars": {
			path: "testdata/nais-valid-template.yaml",
			vars: jsonVars,
		},
		"valid template with yaml vars": {
			path: "testdata/nais-valid-template.yaml",
			vars: yamlVars,
		},
		"valid template with slice vars": {
			path: "testdata/nais-valid-template.yaml",
			vars: sliceVars,
		},
		"valid template with empty vars": {
			path: "testdata/nais-valid-template.yaml",
		},
		"invalid": {
			path:    "testdata/nais-invalid.yaml",
			wantErr: true,
		},
		"invalid multi-document": {
			path:    "testdata/nais-invalid-multidocument.yaml",
			wantErr: true,
		},
		"invalid template with json vars": {
			path:    "testdata/nais-invalid-template.yaml",
			vars:    jsonVars,
			wantErr: true,
		},
		"invalid template with yaml vars": {
			path:    "testdata/nais-invalid-template.yaml",
			vars:    yamlVars,
			wantErr: true,
		},
		"invalid template with slice vars": {
			path:    "testdata/nais-invalid-template.yaml",
			vars:    sliceVars,
			wantErr: true,
		},
		"invalid template with empty vars": {
			path:    "testdata/nais-invalid-template.yaml",
			wantErr: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			v := New([]string{test.path})
			v.SchemaLoader = schemaLoader
			v.Variables = test.vars

			l := naistrix.OutputVerbosityLevelNormal
			err := v.Validate(naistrix.NewOutputWriter(os.Stdout, os.Stderr, &l))
			if test.wantErr && err == nil {
				t.Error("expected a validation error")
			}
			if !test.wantErr && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}
