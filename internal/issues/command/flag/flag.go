package flag

import (
	"context"

	"github.com/nais/cli/internal/alpha/command/flag"
	"github.com/nais/naistrix"
)

type Issues struct {
	*flag.Alpha
}

type List struct {
	*Issues
	Filter   string   `name:"filter" usage:"Filter output (environment,severity,resourceType,resourceName,issueType)=value,..."`
	Severity Severity `name:"severity" usage:"Filter issues by severity (CRITICAL, WARNING, TODO)"`
}
type Severity string

func (s *Severity) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"CRITICAL", "WARNING", "TODO"}, "Available severities"
}
