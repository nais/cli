package formatting

import (
	"fmt"

	"github.com/nais/cli/internal/naisapi/gql"
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
