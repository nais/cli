package validate

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/nais/naistrix"
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

func (v Validate) Validate(out *naistrix.OutputWriter) error {
	invalid := make([]string, 0)

	for _, file := range v.ResourcePaths {
		documents, err := v.loadFile(file, out)
		if err != nil {
			return err
		}

		errors := make([]gojsonschema.ResultError, 0)
		for _, document := range documents {
			documentLoader := gojsonschema.NewBytesLoader(document)
			result, err := gojsonschema.Validate(v.SchemaLoader, documentLoader)
			if err != nil {
				return fmt.Errorf("failed to validate nais manifest: %w", err)
			}

			if !result.Valid() {
				errors = append(errors, result.Errors()...)
			}
		}

		if len(errors) == 0 {
			out.Printf("[✅] %q is valid\n", file)
		} else {
			out.Printf("[❌] %q is invalid\n", file)
			printErrors(errors, out)
			invalid = append(invalid, file)
		}
	}

	if len(invalid) > 0 {
		return fmt.Errorf("validation failed for %d file(s): %s", len(invalid), strings.Join(invalid, ", "))
	}

	return nil
}

func (v Validate) loadFile(name string, out *naistrix.OutputWriter) ([]json.RawMessage, error) {
	_, err := os.Stat(name)
	if err != nil {
		return nil, fmt.Errorf("file %s does not exist", name)
	}

	raw, err := os.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", name, err)
	}

	templated, err := ExecTemplate(raw, v.Variables, out)
	if err != nil {
		return nil, err
	}

	if v.Verbose {
		out.Printf("[🖨️] Printing %q...\n---\n%s", name, templated)
	}

	return YAMLToJSONMessages(templated)
}

func printErrors(errors []gojsonschema.ResultError, out *naistrix.OutputWriter) {
	for _, err := range errors {
		// skip noisy root error ("Must validate one and only one schema (oneOf)")
		if err.Field() == gojsonschema.STRING_ROOT_SCHEMA_PROPERTY && err.Type() == "number_one_of" {
			continue
		}

		out.Printf(" | %q:\n", err.Field())
		out.Printf(" |   - %s\n", err.Description())
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
