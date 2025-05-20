package validate

import (
	"fmt"

	"github.com/nais/cli/internal/root"
)

type Flags struct {
	*root.Flags
	VarsFilePath string
	Vars         []string
}

func Run(files []string, flags *Flags) error {
	templateVars := make(TemplateVariables)

	if flags.VarsFilePath != "" {
		var err error
		templateVars, err = TemplateVariablesFromFile(flags.VarsFilePath)
		if err != nil {
			return fmt.Errorf("load template variables: %v", err)
		}
		for key, val := range templateVars {
			if flags.IsVerbose() {
				fmt.Printf("[📝] Setting template variable '%s' to '%v'\n", key, val)
			}
			templateVars[key] = val
		}
	}

	if len(flags.Vars) > 0 {
		overrides := TemplateVariablesFromSlice(flags.Vars)
		for key, val := range overrides {
			if flags.IsVerbose() {
				if oldval, ok := templateVars[key]; ok {
					fmt.Printf("[⚠️] Overwriting template variable '%s'; previous value was '%v'\n", key, oldval)
				}
				fmt.Printf("[📝] Setting template variable '%s' to '%v'\n", key, val)
			}
			templateVars[key] = val
		}
	}

	v := New(files)
	v.Variables = templateVars
	v.Verbose = flags.IsVerbose()
	return v.Validate()
}
