package auth

import (
	"context"

	"golang.org/x/oauth2"
)

// oidcUser represents a user's session, authenticated via OpenID Connect.
type oidcUser struct {
	oauth2.Token
	IDToken     string `json:"id_token"`
	ConsoleHost string `json:"console_host"`
}

// getOIDCUser holds the lock while reading and refreshing, so concurrent callers refresh only once.
func getOIDCUser(ctx context.Context) (*oidcUser, error) {
	creds, err := lockCredentials(ctx)
	if err != nil {
		return nil, err
	}
	defer creds.unlock()

	user, err := creds.read()
	if err != nil {
		return nil, err
	}

	if user.Valid() {
		return user, nil
	}

	return creds.refresh(ctx, user)
}

// refreshOIDCUser refreshes the session, unless another process has stored a newer one since previous was read.
func refreshOIDCUser(ctx context.Context, previous *oidcUser) (*oidcUser, error) {
	creds, err := lockCredentials(ctx)
	if err != nil {
		return nil, err
	}
	defer creds.unlock()

	stored, err := creds.read()
	if err != nil {
		return nil, err
	}

	if stored.IDToken != previous.IDToken || stored.AccessToken != previous.AccessToken {
		return stored, nil
	}

	return creds.refresh(ctx, stored)
}

func storeOIDCUser(ctx context.Context, tok *oauth2.Token, consoleURL string) (*oidcUser, error) {
	creds, err := lockCredentials(ctx)
	if err != nil {
		return nil, err
	}
	defer creds.unlock()

	return creds.store(tok, consoleURL)
}

// deleteOIDCUser also deletes the shared encryption key, which logs out every config directory.
func deleteOIDCUser(ctx context.Context) error {
	creds, err := lockCredentials(ctx)
	if err != nil {
		return err
	}
	defer creds.unlock()

	return creds.delete()
}
