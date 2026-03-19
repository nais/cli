package command

import (
	"github.com/nais/cli/internal/aiven/command/flag"
	"github.com/nais/naistrix"
)

func grantAccess(parentFlags *flag.Aiven) *naistrix.Command {
	grantAccessFlags := &flag.GrantAccess{Aiven: parentFlags}

	return &naistrix.Command{
		Name:        "grant-access",
		Title:       "Grant a user access to an Aiven service.",
		Description: "Grant a user's service-user access to Aiven Kafka topics or streams.",
		StickyFlags: grantAccessFlags,
		SubCommands: []*naistrix.Command{
			grantAccessStream(grantAccessFlags),
			grantAccessTopic(grantAccessFlags),
		},
	}
}
