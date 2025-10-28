package flag

import (
	"github.com/nais/cli/internal/alpha/command/flag"
)

type Issues struct {
	*flag.Alpha
}

type List struct {
	*Issues
	Filter string `name: "filter", usage:"filter"`
}

type Filters struct {
	IssueType    string
	Severity     string
	Environment  string
	ResourceName string
	ResourceType string
}
