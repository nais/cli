package command

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/command/flag"
	"github.com/nais/naistrix"
)

func proxyCommand(parentFlags *flag.Api) *naistrix.Command {
	flags := &flag.Proxy{Api: parentFlags, ListenAddr: "localhost:4242"}
	return &naistrix.Command{
		Name:        "proxy",
		Title:       "Proxy requests to the Nais API.",
		Description: "This command is used to forward requests to the Nais API, allowing you to interact with the API through a local proxy.",
		Flags:       flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, _ []string) error {
			return naisapi.StartProxy(ctx, out, flags)
		},
	}
}
