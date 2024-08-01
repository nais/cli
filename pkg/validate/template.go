package validate

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
	// We use a fork of aymerick/raymond because the original library does not detect/handle missing template variables.
	"github.com/mailgun/raymond/v2"
)

// Most of this file is copied from https://github.com/nais/deploy/blob/b3ee57a58e6ffbc7dc0586f3781a41e807eda467/pkg/deployclient/template.go for parity.

type TemplateVariables map[string]interface{}

func TemplateVariablesFromFile(path string) (TemplateVariables, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	vars := TemplateVariables{}
	err = yaml.Unmarshal(file, &vars)

	return vars, err
}

func TemplateVariablesFromSlice(vars []string) TemplateVariables {
	tv := TemplateVariables{}
	for _, keyval := range vars {
		tokens := strings.SplitN(keyval, "=", 2)
		switch len(tokens) {
		case 2: // KEY=VAL
			tv[tokens[0]] = tokens[1]
		case 1: // KEY
			tv[tokens[0]] = true
		default:
			continue
		}
	}

	return tv
}

// ExecTemplate evaluates a template with the given context.
func ExecTemplate(data []byte, ctx TemplateVariables) ([]byte, error) {
	if ctx == nil {
		ctx = make(TemplateVariables)
	}

	template, err := raymond.Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse template file: %s", err)
	}

	// extract the base set of expected TemplateVariables
	// string values are in the form of 'test_<variable>'
	templateVars, err := template.ExtractTemplateVars()
	if err != nil {
		return nil, fmt.Errorf("extract template variables: %s", err)
	}

	// if no variables are expected, return the data as is
	if len(templateVars) == 0 && len(ctx) == 0 {
		return data, nil
	}

	// log missing expected variables
	missing := false
	for key, val := range templateVars {
		if _, ok := ctx[key]; !ok {
			missing = true

			b, err := json.Marshal(val)
			if err == nil {
				val = string(b)
			}

			fmt.Printf("[⚠️] Missing template variable for {{%s}}; using placeholder value '%s'\n", key, val)
		}
	}
	if missing {
		fmt.Printf("[⚠️] Placeholder values may be invalid. Provide the missing variables to remove these warnings.\n")
	}

	// override expected variables with values from context
	maps.Copy(templateVars, ctx)

	output, err := template.Exec(templateVars)
	if err != nil {
		return nil, fmt.Errorf("execute template: %s", err)
	}

	return []byte(output), nil
}
