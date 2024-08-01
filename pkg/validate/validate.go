package validate

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/xeipuuv/gojsonschema"
)

const (
	// NaisManifestSchema is the path to the JSON schema for validating a nais manifest
	NaisManifestSchema = "https://storage.googleapis.com/nais-json-schema-2c91/nais-all.json"
)

func init() {
	gojsonschema.Locale = locale{}
}

type Validate struct {
	ResourcePaths []string
	Variables     TemplateVariables
	Verbose       bool
	SchemaLoader  gojsonschema.JSONLoader
}

func New(resourcePaths []string) Validate {
	return Validate{
		ResourcePaths: resourcePaths,
		SchemaLoader:  gojsonschema.NewReferenceLoader(NaisManifestSchema),
	}
}

func (v Validate) Validate() error {
	schema := gojsonschema.NewReferenceLoader(NaisManifestSchema)
	invalid := make([]string, 0)

	for _, file := range v.ResourcePaths {
		document, err := v.loadDocument(file)
		if err != nil {
			return err
		}

		result, err := gojsonschema.Validate(schema, document)
		if err != nil {
			return fmt.Errorf("failed to validate nais manifest: %w", err)
		}

		if result.Valid() {
			fmt.Printf("[✅] %q is valid\n", file)
		} else {
			fmt.Printf("[❌] %q is invalid\n", file)
			printErrors(result.Errors())
			invalid = append(invalid, file)
		}
	}

	if len(invalid) > 0 {
		return fmt.Errorf("validation failed for %d file(s): %s", len(invalid), strings.Join(invalid, ", "))
	}

	return nil
}

func (v Validate) loadDocument(name string) (gojsonschema.JSONLoader, error) {
	_, err := os.Stat(name)
	if err != nil {
		return nil, fmt.Errorf("file %s does not exist", name)
	}

	raw, err := os.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", name, err)
	}

	rendered, err := ExecTemplate(raw, v.Variables)
	if err != nil {
		errMsg := strings.ReplaceAll(err.Error(), "\n", ": ")
		return nil, fmt.Errorf("%s: %s", name, errMsg)
	}

	if v.Verbose {
		fmt.Printf("[🖨️] Printing %q...\n---\n%s", name, rendered)
	}

	var src any
	err = yaml.Unmarshal(rendered, &src)
	if err != nil {
		return nil, fmt.Errorf("parsing yaml: %w", err)
	}

	return gojsonschema.NewGoLoader(src), nil
}

func printErrors(errors []gojsonschema.ResultError) {
	for _, err := range errors {
		// skip noisy root error ("Must validate one and only one schema (oneOf)")
		if err.Field() == gojsonschema.STRING_ROOT_SCHEMA_PROPERTY && err.Type() == "number_one_of" {
			continue
		}

		fmt.Printf(" | %q:\n", err.Field())
		fmt.Printf(" |   - %s\n", err.Description())
	}
}

// locale overrides the default error strings from gojsonschema.DefaultLocale
type locale struct {
	gojsonschema.DefaultLocale
}

func (l locale) AdditionalPropertyNotAllowed() string {
	return `unsupported field "{{.property}}"; it might be misspelled or incorrectly indented. Fields are case sensitive.`
}

func (l locale) Enum() string {
	return `invalid value: must be one of [{{.allowed}}]`
}

func (l locale) InvalidType() string {
	return `invalid type: expected "{{.expected}}", found "{{.given}}"`
}

func (l locale) Required() string {
	return `missing required field "{{.property}}"`
}
