package main

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/cli/internal/root"
)

func main() {
	ctx := context.Background()
	applicationFlags := root.Flags{}

	(&cli.Application{
		Commands: []*cli.Command{
			naisdevice.Command(&applicationFlags),
		},
	}).Run(ctx)

	// if err := cli.Run(context.Background()); err != nil {
	// 	os.Exit(1)
	// }
}
