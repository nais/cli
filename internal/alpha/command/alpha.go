package command

import (
	"github.com/nais/cli/internal/alpha/command/flag"
	apply "github.com/nais/cli/internal/apply/command"
	log "github.com/nais/cli/internal/log/command"
	naisapi "github.com/nais/cli/internal/naisapi/command"
	opensearch "github.com/nais/cli/internal/opensearch/command"
	"github.com/nais/cli/internal/root"
	valkey "github.com/nais/cli/internal/valkey/command"
	"github.com/nais/naistrix"
)

func Alpha(parentFlags *root.Flags) *naistrix.Command {
	flags := &flag.Alpha{Flags: parentFlags}
	return &naistrix.Command{
		Name:        "alpha",
		Title:       "Alpha versions of Nais CLI commands.",
		Description: "These commands are usually fully functional and ready to use, but the API might evolve based on your feedback.",
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			naisapi.Api(flags),
			apply.Apply(flags),
			valkey.Valkey(flags),
			opensearch.OpenSearch(flags),
			log.Log(flags),
		},
	}
}
