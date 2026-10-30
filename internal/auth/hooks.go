package auth

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/gcloud"
	"github.com/nais/cli/internal/naisapi"
	naisapiauth "github.com/nais/cli/internal/naisapi/auth"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/input"
	"golang.org/x/term"
)

const (
	NaisLoginHookName   = "nais-login"
	GcloudLoginHookName = "gcloud-login"
)

// NaisLoginHook authenticates before argument resolution, without rerunning the command.
func NaisLoginHook(f *flags.GlobalFlags) naistrix.PreRunHook {
	return loginHook(NaisLoginHookName, "Nais", naisapi.ErrNeedsLogin, func(ctx context.Context) error {
		_, err := naisapi.GetAuthenticatedUser(ctx)
		return err
	}, naisapi.Login, func() bool { return f.IsVerbose() }, stdinIsTerminal, func() bool { return f.AutoLogin }, confirmLogin)
}

// GcloudLoginHook checks the active gcloud account and Application Default Credentials.
func GcloudLoginHook(f *flags.GlobalFlags) naistrix.PreRunHook {
	return loginHook(GcloudLoginHookName, "Google Cloud", gcloud.ErrNeedsLogin, func(ctx context.Context) error {
		_, err := gcloud.ValidateAndGetUserLogin(ctx, false)
		return err
	}, func(ctx context.Context, out *naistrix.OutputWriter) error {
		return gcloud.Login(ctx, out, f.IsVerbose())
	}, f.IsVerbose, stdinIsTerminal, func() bool { return f.AutoLogin }, confirmLogin)
}

func confirmLogin(prompt string) (bool, error) {
	return input.Confirm(prompt, input.ConfirmWithDefaultTrue())
}

func stdinIsTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) // #nosec G115
}

func loginHook(name, provider string, needsLogin error, check func(context.Context) error, login naistrix.PreRunHookFunc, verbose, interactive, autoLogin func() bool, confirm func(string) (bool, error)) naistrix.PreRunHook {
	return naistrix.PreRunHook{
		Name: name,
		RunFunc: func(ctx context.Context, out *naistrix.OutputWriter) error {
			err := check(ctx)
			if !errors.Is(err, needsLogin) || !interactive() {
				return err
			}

			if rejected, ok := errors.AsType[*naisapiauth.NeedsLoginError](err); ok && verbose() {
				out.Printf("%s authentication required: %s\n", provider, rejected.Cause)
			}
			if !autoLogin() {
				out.Println("Skip prompts: `nais defaults set autologin true`.")
				accepted, promptErr := confirm(fmt.Sprintf("Log in to %s to continue?", provider))
				if promptErr != nil {
					return promptErr
				}
				if !accepted {
					return err
				}
			}
			return login(ctx, out)
		},
	}
}
