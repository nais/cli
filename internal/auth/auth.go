package auth

import (
	"github.com/nais/cli/internal/auth/flag"
	"github.com/nais/cli/internal/auth/login"
	"github.com/nais/cli/internal/auth/logout"
	"github.com/nais/cli/internal/auth/printaccesstoken"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

func Auth(parentFlags *flags.GlobalFlags) *naistrix.Command {
	flags := &flag.Auth{GlobalFlags: parentFlags}
	return &naistrix.Command{
		Name:        "auth",
		Title:       "Authentication",
		Description: "Commands related to authentication in the nais platform",
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			login.Login(flags),
			logout.Logout(flags),
			printaccesstoken.PrintAccessToken(flags),
		},
	}
}
