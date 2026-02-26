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
	Team string `name:"team" short:"t" usage:"Specify the team to use for this command. Overrides the default team from configuration."`
}

func (a AdditionalFlags) RequiredTeam() (string, error) {
	if a.Team == "" {
		return "", fmt.Errorf("team flag is required (use --team flag)")
	}
	return a.Team, nil
}
