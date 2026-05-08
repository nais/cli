package flags

import (
	"fmt"

	"github.com/nais/naistrix"
)

type GlobalFlags struct {
	*naistrix.GlobalFlags
	*AdditionalFlags
}

type AdditionalFlags struct {
	Team        string      `name:"team" short:"t" usage:"Specify the |team| to use for this command. Overrides the default team from configuration."`
	Environment Environment `name:"environment" short:"e" usage:"Specify the |environment| to use for this command. Overrides the default environment from configuration."`
}

func (a AdditionalFlags) RequiredTeam() (string, error) {
	if a.Team == "" {
		return "", fmt.Errorf("team flag is required (use --team flag)")
	}
	return a.Team, nil
}

// HasTeam returns true if the value is not nil and that the [AdditionalFlags.Team] field is not empty.
func (a *AdditionalFlags) HasTeam() bool {
	return a != nil && a.Team != ""
}

// HasEnvironment returns true if the value is not nil and that the [AdditionalFlags.Environment] field is not empty.
func (a *AdditionalFlags) HasEnvironment() bool {
	return a != nil && a.Environment != ""
}
