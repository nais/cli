package krakend

import (
	"context"

	"github.com/nais/cli/internal/alpha/command/flag"
	"github.com/nais/krakend/pkg/migration"
	"github.com/nais/naistrix"
)

func Krakend(parentFlags *flag.Alpha) *naistrix.Command {
	return &naistrix.Command{
		Name:  "krakend",
		Title: "Krakend related functionality.",
		SubCommands: []*naistrix.Command{
			{
				Name:        "convert",
				Title:       "Fetch and convert krakend resources to YAML.",
				Description: "Temporary command to convert all Krakend resources in current namespace to Application and relevant config maps",
				RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
					ret, err := migration.ConvertKrakends(ctx)
					if err != nil {
						return err
					}
					out.Println(ret)
					return nil
				},
			},
		},
	}
}
