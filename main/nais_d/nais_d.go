package main

import "github.com/nais/nais-d/cmd"

var (
	// Is set during build
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	cmd.Execute(version, commit, date, builtBy)
}
