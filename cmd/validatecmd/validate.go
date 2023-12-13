package validatecmd

import (
	"fmt"
	"github.com/nais/cli/pkg/validate"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:            "validate",
		Usage:           "Validate nais.yaml configuration",
		ArgsUsage:       "nais.yaml [naiser.yaml...]",
		UsageText:       "nais validate nais.yaml [naiser.yaml...]",
		HideHelpCommand: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "vars",
				Usage: "path to `FILE` containing template variables, must be JSON or YAML format.",
			},
			&cli.StringSliceFlag{
				Name:    "var",
				Aliases: []string{"v"},
				Usage:   "template variable in KEY=VALUE form, can be specified multiple times.",
			},
		},
		Before: func(context *cli.Context) error {
			if context.Args().Len() == 0 {
				return fmt.Errorf("no config files provided")
			}

			return nil
		},
		Action: func(context *cli.Context) error {
			resources := context.Args().Slice()
			varsPath := context.String("vars")
			vars := context.StringSlice("var")

			templateVars := make(validate.TemplateVariables)
			var err error

			if varsPath != "" {
				templateVars, err = validate.TemplateVariablesFromFile(varsPath)
				if err != nil {
					return fmt.Errorf("load template variables: %v", err)
				}
				for key, val := range templateVars {
					fmt.Printf("Setting template variable '%s' to '%v'\n", key, val)
					templateVars[key] = val
				}
			}

			if len(vars) > 0 {
				overrides := validate.TemplateVariablesFromSlice(vars)
				for key, val := range overrides {
					if oldval, ok := templateVars[key]; ok {
						fmt.Printf("Overwriting template variable '%s'; previous value was '%v'\n", key, oldval)
					}
					fmt.Printf("Setting template variable '%s' to '%v'\n", key, val)
					templateVars[key] = val
				}
			}

			return validate.NaisConfig(resources, templateVars)
		},
	}
}
