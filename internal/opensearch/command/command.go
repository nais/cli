package command

import (
	alpha "github.com/nais/cli/internal/alpha/command/flag"
	"github.com/nais/cli/internal/opensearch/command/flag"
	"github.com/nais/naistrix"
)

func OpenSearch(parentFlags *alpha.Alpha) *naistrix.Command {
	flags := &flag.OpenSearch{Alpha: parentFlags}
	return &naistrix.Command{
		Name:        "opensearch",
		Title:       "Manage OpenSearch instances.",
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			createOpenSearch(flags),
			deleteOpenSearch(flags),
			describeOpenSearch(flags),
			listOpenSearches(flags),
			updateOpenSearch(flags),
		},
	}
}
