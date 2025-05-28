package validate

import (
	_ "embed"
	"testing"

	"github.com/nais/cli/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
)

//go:embed schema.json
var schema []byte

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

			err := v.Validate(output.Stdout())
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
