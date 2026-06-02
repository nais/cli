package command

import (
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/naistrix"
)

func set(parentFlags *flag.App) *naistrix.Command {
	return &naistrix.Command{
		Name:        "set",
		Title:       "Update application configuration.",
		Description: "Commands for updating application settings such as environment variables and replicas. Changes are temporary and will be overwritten on next deploy.",
		SubCommands: []*naistrix.Command{
			setReplicas(parentFlags),
			setEnv(parentFlags),
			setImage(parentFlags),
		},
	}
}
