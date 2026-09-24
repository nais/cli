package validate

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/nais/naistrix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			require.NoError(t, os.WriteFile(path, []byte(manifest), 0o600))

			v := New([]string{path})
			v.SchemaLoader = gojsonschema.NewBytesLoader(schema)
			level := naistrix.OutputVerbosityLevelNormal
			err := v.Validate(naistrix.NewOutputWriter(io.Discard, &level))
			require.ErrorContains(t, err, tc.kind+" manifests cannot be validated with nais validate")
			assert.ErrorContains(t, err, "use nais apply --dry-run instead")
		})
	}
}

func TestValidate(t *testing.T) {
	schemaLoader := gojsonschema.NewBytesLoader(schema)

	jsonVars, err := TemplateVariablesFromFile("testdata/vars.json")
	require.NoError(t, err)
	require.NotEmpty(t, jsonVars)

	yamlVars, err := TemplateVariablesFromFile("testdata/vars.yaml")
	require.NoError(t, err)
	require.NotEmpty(t, yamlVars)

	sliceVars := TemplateVariablesFromSlice([]string{
		"app=some-app",
		"namespace=some-namespace",
		"image=some-image",
		"team=some-team",
	})
	require.Contains(t, sliceVars, "app")
	require.Contains(t, sliceVars, "namespace")
	require.Contains(t, sliceVars, "image")
	require.Contains(t, sliceVars, "team")

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
			err := v.Validate(naistrix.NewOutputWriter(os.Stdout, &l))
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
