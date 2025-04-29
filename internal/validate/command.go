package validate

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/metrics"
	"github.com/urfave/cli/v3"
)

func Before(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	if cmd.Args().Len() == 0 {
		metrics.AddOne(ctx, "validate_enonent_error_total")
		return ctx, fmt.Errorf("no config files provided")
	}

	return ctx, nil
}

func Action(ctx context.Context, cmd *cli.Command) error {
	resourcePaths := cmd.Args().Slice()
	varsPath := cmd.String("vars")
	vars := cmd.StringSlice("var")
	verbose := cmd.Bool("verbose")

	templateVars := make(TemplateVariables)
	var err error

	if varsPath != "" {
		templateVars, err = TemplateVariablesFromFile(varsPath)
		if err != nil {
			return fmt.Errorf("load template variables: %v", err)
		}
		for key, val := range templateVars {
			if verbose {
				fmt.Printf("[📝] Setting template variable '%s' to '%v'\n", key, val)
			}
			templateVars[key] = val
		}
	}

	if len(vars) > 0 {
		overrides := TemplateVariablesFromSlice(vars)
		for key, val := range overrides {
			if verbose {
				if oldval, ok := templateVars[key]; ok {
					fmt.Printf("[⚠️] Overwriting template variable '%s'; previous value was '%v'\n", key, oldval)
				}
				fmt.Printf("[📝] Setting template variable '%s' to '%v'\n", key, val)
			}
			templateVars[key] = val
		}
	}

	v := New(resourcePaths)
	v.Variables = templateVars
	v.Verbose = verbose
	return v.Validate()
}
