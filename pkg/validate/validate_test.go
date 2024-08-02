package validate

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
)

//go:embed schema.json
var schema []byte

func TestValidate(t *testing.T) {
	t.Run("non-templated yaml", func(t *testing.T) {
		v := New([]string{"testdata/nais-valid.yaml"})
		v.SchemaLoader = gojsonschema.NewBytesLoader(schema)
		err := v.Validate()
		assert.NoError(t, err)

		v = New([]string{"testdata/nais-invalid.yaml"})
		v.SchemaLoader = gojsonschema.NewBytesLoader(schema)
		err = v.Validate()
		assert.Error(t, err)
	})

	t.Run("multi-document yaml", func(t *testing.T) {
		v := New([]string{"testdata/nais-valid-multidocument.yaml"})
		v.SchemaLoader = gojsonschema.NewBytesLoader(schema)
		err := v.Validate()
		assert.NoError(t, err)

		v = New([]string{"testdata/nais-invalid-multidocument.yaml"})
		v.SchemaLoader = gojsonschema.NewBytesLoader(schema)
		err = v.Validate()
		assert.Error(t, err)
	})

	t.Run("templated yaml", func(t *testing.T) {
		t.Run("variables from file", func(t *testing.T) {
			for _, file := range []string{"testdata/vars.json", "testdata/vars.yaml"} {
				t.Run(file, func(t *testing.T) {
					vars, err := TemplateVariablesFromFile(file)
					assert.NoError(t, err)

					v := New([]string{"testdata/nais-valid-template.yaml"})
					v.Variables = vars
					v.SchemaLoader = gojsonschema.NewBytesLoader(schema)
					err = v.Validate()
					assert.NoError(t, err)

					v = New([]string{"testdata/nais-invalid-template.yaml"})
					v.Variables = vars
					v.SchemaLoader = gojsonschema.NewBytesLoader(schema)
					err = v.Validate()
					assert.Error(t, err)
				})
			}
		})

		t.Run("variables from slice", func(t *testing.T) {
			vars := TemplateVariablesFromSlice([]string{
				"app=some-app",
				"namespace=some-namespace",
				"image=some-image",
				"team=some-team",
			})

			v := New([]string{"testdata/nais-valid-template.yaml"})
			v.Variables = vars
			v.SchemaLoader = gojsonschema.NewBytesLoader(schema)
			err := v.Validate()
			assert.NoError(t, err)

			v = New([]string{"testdata/nais-invalid-template.yaml"})
			v.Variables = vars
			v.SchemaLoader = gojsonschema.NewBytesLoader(schema)
			err = v.Validate()
			assert.Error(t, err)
		})

		t.Run("no variables provided", func(t *testing.T) {
			v := New([]string{"testdata/nais-valid-template.yaml"})
			err := v.Validate()
			assert.NoError(t, err)

			v = New([]string{"testdata/nais-invalid-template.yaml"})
			err = v.Validate()
			assert.Error(t, err)
		})
	})
}
