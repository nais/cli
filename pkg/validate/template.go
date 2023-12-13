package validate

import (
	"fmt"
	"os"
	"strings"

	"github.com/aymerick/raymond"
	"github.com/ghodss/yaml"
)

type TemplateVariables map[string]interface{}

func templatedFile(data []byte, ctx TemplateVariables) ([]byte, error) {
	if len(ctx) == 0 {
		return data, nil
	}
	template, err := raymond.Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse template file: %s", err)
	}

	output, err := template.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("execute template: %s", err)
	}

	return []byte(output), nil
}

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
