package auth

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestConcurrentCallersRefreshOnce(t *testing.T) {
	setupCredentialStore(t)
	tok := startTestOIDCProvider(t, "nav.no").token(t, time.Hour)
	tok.Expiry = time.Now().Add(-time.Hour)
	if _, err := storeOIDCUser(t.Context(), tok, "console.nav.cloud.nais.io"); err != nil {
		t.Fatal(err)
	}
	requests := countTokenRequests(t)

	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for range cap(errs) {
		wg.Go(func() {
			user, err := getOIDCUser(t.Context())
			if err == nil && user.AccessToken != "new-access-token" {
				err = fmt.Errorf("access token = %q, want refreshed token", user.AccessToken)
			}
			errs <- err
		})
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if got := requests.Load(); got != 1 {
		t.Fatalf("token endpoint requests = %d, want 1", got)
	}
}
