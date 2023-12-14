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

type Validate struct {
	ResourcePaths []string
	Variables     TemplateVariables
	Verbose       bool
}

func New(resourcePaths []string) Validate {
	return Validate{
		ResourcePaths: resourcePaths,
	}
}

func (v Validate) Validate() error {
	fail := make([]string, 0)

	for _, file := range v.ResourcePaths {
		if _, err := os.Stat(file); err != nil {
			return fmt.Errorf("file %s does not exist", file)
		}

		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", file, err)
		}

		content, err = templatedFile(content, v.Variables)
		if err != nil {
			errMsg := strings.ReplaceAll(err.Error(), "\n", ": ")
			return fmt.Errorf("%s: %s", file, errMsg)
		}

		if v.Verbose {
			fmt.Printf("[🖨️] Printing %q...\n", file)
			fmt.Println("---")
			fmt.Println(string(content))
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
			fmt.Printf("[✅] %q is valid\n", file)
		} else {
			fmt.Printf("[❌] %q is not valid and has the following errors:\n", file)
			for _, desc := range result.Errors() {
				fmt.Printf("- %s\n", desc)
			}
			fail = append(fail, file)
		}
	}

	if len(fail) > 0 {
		return fmt.Errorf("validation failed for %d file(s): %s", len(fail), strings.Join(fail, ", "))
	}

	return nil
}
