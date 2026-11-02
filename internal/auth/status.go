package auth

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/nais/cli/internal/gcloud"
	"github.com/nais/cli/internal/naisapi"
)

// LoginState is a stable machine-readable result of a provider's credential check.
type LoginState string

const (
	LoginStateAuthenticated LoginState = "authenticated"
	LoginStateNotLoggedIn   LoginState = "not_logged_in"
	LoginStateNotInstalled  LoginState = "not_installed"
	LoginStateCheckFailed   LoginState = "check_failed"
)

func (s LoginState) String() string {
	switch s {
	case LoginStateAuthenticated:
		return "Authenticated"
	case LoginStateNotLoggedIn:
		return "Not logged in"
	case LoginStateNotInstalled:
		return "Not installed"
	case LoginStateCheckFailed:
		return "Check failed"
	default:
		return string(s)
	}
}

// LoginStatus describes whether a provider's credentials satisfy the CLI's login check.
type LoginStatus struct {
	Provider string     `json:"-"`
	Status   LoginState `json:"status"`
	Account  *string    `json:"account"`
	Details  string     `json:"details,omitempty"`
}

type LoginStatusReport struct {
	Nais   LoginStatus `json:"nais"`
	Gcloud LoginStatus `json:"gcloud"`
}

type loginStatusCheck struct {
	provider   string
	needsLogin error
	check      func(context.Context) (string, error)
}

// LoginStatuses checks both providers without starting an interactive login flow.
// Missing credentials and a missing gcloud executable are reported as statuses, not errors.
// Unexpected errors are returned after both providers have been checked.
func LoginStatuses(ctx context.Context) (LoginStatusReport, error) {
	statuses, err := checkLoginStatuses(ctx, []loginStatusCheck{
		{
			provider: "Nais", needsLogin: naisapi.ErrNeedsLogin,
			check: func(ctx context.Context) (string, error) {
				user, err := naisapi.GetAuthenticatedUser(ctx)
				if err != nil {
					return "", err
				}
				return user.Email(), nil
			},
		},
		{
			provider: "Google Cloud", needsLogin: gcloud.ErrNeedsLogin,
			check: func(ctx context.Context) (string, error) {
				return gcloud.ValidateAndGetUserLogin(ctx, false)
			},
		},
	})
	return LoginStatusReport{Nais: statuses[0], Gcloud: statuses[1]}, err
}

func checkLoginStatuses(ctx context.Context, checks []loginStatusCheck) ([]LoginStatus, error) {
	statuses := make([]LoginStatus, 0, len(checks))
	var errs []error
	for _, check := range checks {
		status := LoginStatus{Provider: check.provider}
		account, err := check.check(ctx)
		switch {
		case err == nil:
			status.Status = LoginStateAuthenticated
			if account != "" {
				status.Account = &account
			}
		case errors.Is(err, check.needsLogin):
			status.Status = LoginStateNotLoggedIn
			status.Details = err.Error()
		case check.provider == "Google Cloud" && errors.Is(err, exec.ErrNotFound):
			status.Status = LoginStateNotInstalled
			status.Details = "Install gcloud and run `nais login`."
		default:
			status.Status = LoginStateCheckFailed
			status.Details = err.Error()
			errs = append(errs, fmt.Errorf("%s authentication status: %w", check.provider, err))
		}
		statuses = append(statuses, status)
	}
	return statuses, errors.Join(errs...)
}
