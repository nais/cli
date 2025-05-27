package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/command/flag"
	"github.com/nais/cli/internal/output"
)

func proxy(parentFlags *flag.Api) *cli.Command {
	flags := &flag.Proxy{Api: parentFlags, ListenAddr: "localhost:4242"}
	return cli.NewCommand("proxy", "Proxy requests to the Nais API.",
		cli.WithLong("This command is used to forward requests to the Nais API, allowing you to interact with the API through a local proxy."),
		cli.WithFlag("listen", "l", "Address the proxy will listen on.", &flags.ListenAddr),
		cli.WithRun(func(ctx context.Context, out output.Output, _ []string) error {
			return naisapi.StartProxy(ctx, out, flags)
		}),
	)
}
