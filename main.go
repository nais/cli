package main

import (
	"context"
	"os"

	"github.com/nais/cli/internal/auth/login"
	"github.com/nais/cli/internal/auth/logout"
	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/cli/internal/root"
)

func main() {
	flags := &root.Flags{}
	app := cli.NewApplication(
		login.Command(flags),
		logout.Command(flags),
		naisdevice.Command(flags),
	)
	if err := app.Run(context.Background()); err != nil {
		// TODO: output error
		os.Exit(1)
	}
}
