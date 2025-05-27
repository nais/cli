package main

import (
	"context"
	"os"

	aiven "github.com/nais/cli/internal/aiven/command"
	alpha "github.com/nais/cli/internal/alpha/command"
	"github.com/nais/cli/internal/auth/login"
	"github.com/nais/cli/internal/auth/logout"
	"github.com/nais/cli/internal/cli"
	debug "github.com/nais/cli/internal/debug/command"
	kubeconfig "github.com/nais/cli/internal/kubeconfig/command"
	naisdevice "github.com/nais/cli/internal/naisdevice/command"
	postgres "github.com/nais/cli/internal/postgres/command"
	"github.com/nais/cli/internal/root"
	validate "github.com/nais/cli/internal/validate/command"
)

func main() {
	flags := &root.Flags{}
	app := cli.NewApplication(flags,
		login.Command(flags),
		logout.Command(flags),
		naisdevice.Naisdevice(flags),
		aiven.Aiven(flags),
		alpha.Alpha(flags),
		postgres.Postgres(flags),
		debug.Debug(flags),
		kubeconfig.Kubeconfig(flags),
		validate.Validate(flags),
	)
	if err := app.Run(context.Background()); err != nil {
		// TODO: output error
		os.Exit(1)
	}
}
