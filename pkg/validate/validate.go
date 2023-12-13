package validate

import (
	"fmt"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/xeipuuv/gojsonschema"
)

const (
	// NaisManifestSchema is the path to the JSON schema for validating a nais manifest
	NaisManifestSchema = "https://storage.googleapis.com/nais-json-schema-2c91/nais-all.json"
)

func NaisConfig(resources []string, variables TemplateVariables) error {
	validationFailed := false

	for _, file := range resources {
		if _, err := os.Stat(file); err != nil {
			return fmt.Errorf("file %s does not exist", file)
		}

		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", file, err)
		}

		content, err = templatedFile(content, variables)
		if err != nil {
			errMsg := strings.ReplaceAll(err.Error(), "\n", ": ")
			return fmt.Errorf("%s: %s", file, errMsg)
		}

		var m interface{}
		err = yaml.Unmarshal(content, &m)
		if err != nil {
			return fmt.Errorf("failed to convert yaml to json: %w", err)
		}

		schemaLoader := gojsonschema.NewReferenceLoader(NaisManifestSchema)
		documentLoader := gojsonschema.NewGoLoader(m)

		result, err := gojsonschema.Validate(schemaLoader, documentLoader)
		if err != nil {
			return fmt.Errorf("failed to validate nais manifest: %w", err)
		}

		if result.Valid() {
			fmt.Printf("%s is valid\n", file)
		} else {
			validationFailed = true

			fmt.Printf("%s is not valid and has the following errors:\n", file)
			for _, desc := range result.Errors() {
				fmt.Printf("- %s\n", desc)
			}
		}
	}

	if validationFailed {
		return fmt.Errorf("validation failed")
	}

	return nil
}
