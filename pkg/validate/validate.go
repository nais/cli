package validate

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	"github.com/xeipuuv/gojsonschema"
)

const (
	// NaisManifestSchema is the path to the JSON schema for validating a nais manifest
	NaisManifestSchema = "https://storage.googleapis.com/nais-json-schema-2c91/nais-all.json"
)

func NaisConfig(config []string) error {
	for _, file := range config {
		if _, err := os.Stat(file); err != nil {
			return fmt.Errorf("file %s does not exist", file)
		}

		content, err := ioutil.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", file, err)
		}

		json, err := yaml.YAMLToJSON(content)
		if err != nil {
			return fmt.Errorf("failed to convert yaml to json: %w", err)
		}

		schemaLoader := gojsonschema.NewReferenceLoader(NaisManifestSchema)
		documentLoader := gojsonschema.NewStringLoader(string(json))

		result, err := gojsonschema.Validate(schemaLoader, documentLoader)
		if err != nil {
			panic(err.Error())
		}

		if result.Valid() {
			fmt.Printf("%s is valid\n", file)
		} else {
			fmt.Printf("%s is not valid and has the following errors:\n", file)
			for _, desc := range result.Errors() {
				fmt.Printf("- %s\n", desc)
			}
		}
	}

	return nil
}
