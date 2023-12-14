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
				Name:  "var",
				Usage: "template variable in KEY=VALUE form, can be specified multiple times.",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "print all the template variables and final resources after templating.",
			},
		},
		Before: func(context *cli.Context) error {
			if context.Args().Len() == 0 {
				return fmt.Errorf("no config files provided")
			}

			return nil
		},
		Action: func(context *cli.Context) error {
			resourcePaths := context.Args().Slice()
			varsPath := context.String("vars")
			vars := context.StringSlice("var")
			verbose := context.Bool("verbose")

			templateVars := make(validate.TemplateVariables)
			var err error

			if varsPath != "" {
				templateVars, err = validate.TemplateVariablesFromFile(varsPath)
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
				overrides := validate.TemplateVariablesFromSlice(vars)
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

			v := validate.New(resourcePaths)
			v.Variables = templateVars
			v.Verbose = verbose
			return v.Validate()
		},
	}
}
