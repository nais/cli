package command

import (
	"github.com/nais/cli/internal/job/command/flag"
	"github.com/nais/naistrix"
)

func run(parentFlags *flag.Job) *naistrix.Command {
	return &naistrix.Command{
		Name:  "run",
		Title: "Manage job runs.",
		SubCommands: []*naistrix.Command{
			listRuns(parentFlags),
			deleteRun(parentFlags),
		},
	}
}
