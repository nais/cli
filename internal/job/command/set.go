package command

import (
	"github.com/nais/cli/internal/job/command/flag"
	"github.com/nais/naistrix"
)

func set(parentFlags *flag.Job) *naistrix.Command {
	return &naistrix.Command{
		Name:        "set",
		Title:       "Update job configuration.",
		Description: "Commands for updating job settings such as environment variables. Changes are temporary and will be overwritten on next deploy.",
		SubCommands: []*naistrix.Command{
			setEnv(parentFlags),
		},
	}
}
