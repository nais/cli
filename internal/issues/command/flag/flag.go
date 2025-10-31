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

//TODO: maybe have some filter definition stuff here we can use to make the parsing generic as well as provide help texts for the filter flag?
