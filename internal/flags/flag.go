package flags

import "github.com/nais/naistrix"

type GlobalFlags struct {
	*naistrix.GlobalFlags
	*AdditionalFlags
}

type AdditionalFlags struct {
	Team string `name:"team" usage:"Specify the team to use for this command. Overrides the default team from configuration."`
}
