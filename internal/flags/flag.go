package flags

import (
	"github.com/nais/naistrix"
)

type GlobalFlags struct {
	*naistrix.GlobalFlags
	*AdditionalFlags
}

type AdditionalFlags struct {
	ConfirmLogin bool        `name:"confirm-login" usage:"Ask for confirmation before starting required login flows. Interactive terminals only."`
	Team         string      `name:"team" short:"t" usage:"Select the Nais |team| that owns the resources. Overrides the team configured with nais defaults set team."`
	Environment  Environment `name:"environment" short:"e" usage:"Select or filter by Nais |environment|, for example dev-gcp or prod-gcp. Overrides the environment configured with nais defaults set environment."`
}

// HasTeam returns true if the value is not nil and that the [AdditionalFlags.Team] field is not empty.
func (a *AdditionalFlags) HasTeam() bool {
	return a != nil && a.Team != ""
}

// GetTeam returns the team value, or an empty string if the receiver is nil.
func (a *AdditionalFlags) GetTeam() string {
	if a == nil {
		return ""
	}
	return a.Team
}

// HasEnvironment returns true if the value is not nil and that the [AdditionalFlags.Environment] field is not empty.
func (a *AdditionalFlags) HasEnvironment() bool {
	return a != nil && a.Environment != ""
}
