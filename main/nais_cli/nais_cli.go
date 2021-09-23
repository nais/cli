package main

import (
	"github.com/nais/nais-cli/cmd/root"
)

var (
	// Is set during build
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	root.Execute(version, commit, date, builtBy)
}
