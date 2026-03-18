package formatting

import (
	"fmt"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/savioxavier/termlink"
)

func ColoredSeverityString(s string, severity gql.Severity) string {
	level := "info"

	switch severity {
	case gql.SeverityCritical:
		level = "error"
	case gql.SeverityWarning:
		level = "warn"
	}

	return fmt.Sprintf("<%v>%v</%v>", level, s, level)
}

func Link(title, url string) string {
	if url == "" || !termlink.SupportsHyperlinks() {
		return title
	}

	return termlink.Link(title, url)
}
