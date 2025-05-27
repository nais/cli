package main

import (
	"context"
	"os"

	aiven "github.com/nais/cli/internal/aiven/command"
	alpha "github.com/nais/cli/internal/alpha/command"
	"github.com/nais/cli/internal/auth/login"
	"github.com/nais/cli/internal/auth/logout"
	"github.com/nais/cli/internal/cli"
	naisdevice "github.com/nais/cli/internal/naisdevice/command"
	"github.com/nais/cli/internal/root"
)

func main() {
	flags := &root.Flags{}
	app := cli.NewApplication(flags,
		login.Command(flags),
		logout.Command(flags),
		naisdevice.Naisdevice(flags),
		aiven.Aiven(flags),
		alpha.Alpha(flags),
	)
	if err := app.Run(context.Background()); err != nil {
		// TODO: output error
		os.Exit(1)
	}
}
