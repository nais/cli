package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/gcloud"
	"github.com/nais/cli/internal/naisapi"
	naisapiauth "github.com/nais/cli/internal/naisapi/auth"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/input"
)

const (
	NaisLoginHookName   = "nais-login"
	GcloudLoginHookName = "gcloud-login"
)

// NaisLoginHook authenticates before argument resolution, without rerunning the command.
func NaisLoginHook(f *flags.GlobalFlags) naistrix.PreRunHook {
	return loginHook(
		NaisLoginHookName,
		"Nais",
		naisapi.ErrNeedsLogin,
		naisapi.CheckAuthentication,
		naisapi.Login,
		func() bool { return f.IsVerbose() },
		input.IsInteractive,
		func() bool { return f.ConfirmLogin },
		confirmLogin,
	)
}

// GcloudLoginHook checks the active gcloud account and Application Default Credentials.
func GcloudLoginHook(f *flags.GlobalFlags) naistrix.PreRunHook {
	return loginHook(
		GcloudLoginHookName,
		"Google Cloud",
		gcloud.ErrNeedsLogin,
		func(ctx context.Context) error {
			_, err := gcloud.ValidateAndGetUserLogin(ctx, false)
			return err
		},
		func(ctx context.Context, out *naistrix.OutputWriter) error {
			return gcloud.Login(ctx, out, f.IsVerbose())
		},
		f.IsVerbose,
		input.IsInteractive,
		func() bool { return f.ConfirmLogin },
		confirmLogin,
	)
}

func confirmLogin(ctx context.Context, prompt string) (bool, error) {
	return input.Confirm(ctx, prompt, input.ConfirmWithDefaultTrue())
}

// loginHook builds a pre-run hook that checks credentials and starts a login flow only if credentials are needed and
// the command can interact with the user.
//
//   - name identifies the hook to naistrix (for example, in SkipPreRunHooks).
//   - provider is the display name used in messages and prompts.
//   - needsLogin is the error that signals login is required, matched with errors.Is.
//   - check validates existing credentials. A nil result lets the command proceed;
//     errors other than needsLogin stop the command without prompting.
//   - login starts the provider's login flow. Its result determines whether the command proceeds.
//   - verbose reports whether to show the authentication failure's cause, when available.
//   - interactive reports whether user interaction is allowed. If false, the hook returns
//     the original check error without prompting or logging in.
//   - shouldConfirm reports whether to ask before starting login. Disabling confirmation
//     skips the prompt, not the login flow itself.
//   - confirm asks for permission to start login. A declined prompt returns the original
//     check error; a prompt failure returns the prompt error.
func loginHook(
	name, provider string,
	needsLogin error,
	check func(context.Context) error,
	login naistrix.PreRunHookFunc,
	verbose func() bool,
	interactive func(context.Context) bool,
	shouldConfirm func() bool,
	confirm func(context.Context, string) (bool, error),
) naistrix.PreRunHook {
	return naistrix.PreRunHook{
		Name: name,
		RunFunc: func(ctx context.Context, out *naistrix.OutputWriter) error {
			err := check(ctx)
			if !errors.Is(err, needsLogin) || !interactive(ctx) {
				return err
			}

			if rejected, ok := errors.AsType[*naisapiauth.NeedsLoginError](err); ok && verbose() {
				out.Printf("%s authentication required: %s\n", provider, rejected.Cause)
			}
			if shouldConfirm() {
				out.Println("Skip prompts: `nais defaults set confirm-login false`.")
				accepted, promptErr := confirm(ctx, fmt.Sprintf("Log in to %s to continue?", provider))
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
