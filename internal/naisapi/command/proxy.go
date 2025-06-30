package command

import (
	"context"

	"github.com/nais/cli/pkg/cli/v2"
	"github.com/nais/cli/v2/internal/naisapi"
	"github.com/nais/cli/v2/internal/naisapi/command/flag"
)

func proxy(parentFlags *flag.Api) *cli.Command {
	flags := &flag.Proxy{Api: parentFlags, ListenAddr: "localhost:4242"}
	return &cli.Command{
		Name:        "proxy",
		Title:       "Proxy requests to the Nais API.",
		Description: "This command is used to forward requests to the Nais API, allowing you to interact with the API through a local proxy.",
		Flags:       flags,
		RunFunc: func(ctx context.Context, out cli.Output, _ []string) error {
			return naisapi.StartProxy(ctx, out, flags)
		},
	}
}
