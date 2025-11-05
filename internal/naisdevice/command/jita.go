package command

import (
	"context"

	"github.com/nais/naistrix"
)

func jitacmd() *naistrix.Command {
	// TODO: remove in a future release
	return &naistrix.Command{
		Name:  "jita",
		Title: "Connect to a JITA gateway.",
		Args: []naistrix.Argument{
			{Name: "gateway", Repeatable: true},
		},
		Deprecated: naistrix.DeprecatedWithReplacementFunc(func(_ context.Context, args *naistrix.Arguments) []string {
			return append([]string{"device", "gateway", "grant-access"}, args.All()...)
		}),
	}
}
