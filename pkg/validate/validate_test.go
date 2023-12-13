package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNaisConfig(t *testing.T) {
	t.Run("non-templated yaml", func(t *testing.T) {
		err := NaisConfig([]string{"testdata/nais-valid.yaml"}, nil)
		assert.NoError(t, err)

		err = NaisConfig([]string{"testdata/nais-invalid.yaml"}, nil)
		assert.Error(t, err)
	})

	t.Run("templated yaml", func(t *testing.T) {
		for _, file := range []string{"testdata/vars.json", "testdata/vars.yaml"} {
			t.Run(file, func(t *testing.T) {
				vars, err := TemplateVariablesFromFile(file)
				assert.NoError(t, err)

				err = NaisConfig([]string{"testdata/nais-valid-template.yaml"}, vars)
				assert.NoError(t, err)

				err = NaisConfig([]string{"testdata/nais-invalid-template.yaml"}, vars)
				assert.Error(t, err)
			})
		}

		strings := []string{
			"app=some-app",
			"namespace=some-namespace",
			"image=some-image",
			"team=some-team",
		}
		vars := TemplateVariablesFromSlice(strings)

		err := NaisConfig([]string{"testdata/nais-valid-template.yaml"}, vars)
		assert.NoError(t, err)

		err = NaisConfig([]string{"testdata/nais-invalid-template.yaml"}, vars)
		assert.Error(t, err)
	})
}
