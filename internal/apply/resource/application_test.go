package resource

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
	"golang.org/x/oauth2"
)

type applicationStatusTransport func(*http.Request) (*http.Response, error)

func (f applicationStatusTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestWait_AppliedMinimum(t *testing.T) {
	for name, tc := range map[string]struct {
		spec       string
		ready      bool
		fresh      bool
		wantWait   bool
		wantOutput string
	}{
		"explicit zero skips waiting even when API still has instances": {
			spec:       "replicas: {min: 0}",
			ready:      true,
			wantOutput: "no instances expected",
		},
		"increased minimum does not trust stale zero minimum": {
			spec:     "replicas: {min: 1}",
			wantWait: true,
		},
		"omitted minimum does not default to zero": {
			spec:     "replicas: {max: 2}",
			wantWait: true,
		},
		"omitted replicas does not default to zero": {
			spec:     "image: example:latest",
			wantWait: true,
		},
		"stale zero minimum does not bypass grace window": {
			spec:     "replicas: {min: 1}",
			ready:    true,
			wantWait: true,
		},
		"increased minimum succeeds after rollout becomes ready": {
			spec:       "replicas: {min: 1}",
			ready:      true,
			fresh:      true,
			wantOutput: "Application/testapp: ready",
		},
	} {
		t.Run(name, func(t *testing.T) {
			docs, err := Documents([]byte("spec:\n  " + tc.spec + "\n"))
			if err != nil {
				t.Fatal(err)
			}

			since := time.Now()
			requests := 0
			client := &http.Client{Transport: applicationStatusTransport(func(_ *http.Request) (*http.Response, error) {
				requests++
				state := "NOT_RUNNING"
				groups := []map[string]any{}
				if tc.ready {
					state = "RUNNING"
					created := since.Add(-time.Hour)
					if tc.fresh {
						created = since.Add(time.Second)
					}
					groups = append(groups, map[string]any{
						"created":          created.Format(time.RFC3339Nano),
						"readyInstances":   1,
						"desiredInstances": 1,
					})
				}
				body, err := json.Marshal(map[string]any{
					"data": map[string]any{
						"team": map[string]any{
							"environment": map[string]any{
								"application": map[string]any{
									"state":          state,
									"resources":      map[string]any{"scaling": map[string]any{"minInstances": 0}},
									"instanceGroups": groups,
								},
							},
						},
					},
				})
				if err != nil {
					return nil, err
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": {"application/json"}},
					Body:       io.NopCloser(bytes.NewReader(body)),
				}, nil
			})}
			t.Setenv("NAIS_API_LOCAL_HOST", "api.test")
			t.Setenv("NAIS_API_LOCAL_EMAIL", "")

			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()
			ctx = context.WithValue(ctx, oauth2.HTTPClient, client)
			var output bytes.Buffer
			err = (applicationResource{}).Wait(ctx, "team", "dev", "testapp", docs[0], since, naistrix.NewOutputWriter(&output, new(naistrix.Count)))

			if tc.wantWait {
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("expected to keep waiting, got %v", err)
				}
				if requests == 0 {
					t.Fatal("expected to poll application status")
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if !tc.fresh && requests != 0 {
					t.Errorf("expected no status requests for explicit min=0, got %d", requests)
				}
				if !strings.Contains(output.String(), tc.wantOutput) {
					t.Errorf("output %q does not contain %q", output.String(), tc.wantOutput)
				}
			}
		})
	}
}

func TestIsConverged(t *testing.T) {
	now := time.Now()

	for name, tc := range map[string]struct {
		state  gql.ApplicationState
		groups []instanceGroup
		want   bool
	}{
		"single fully-ready group while running is converged": {
			state:  gql.ApplicationStateRunning,
			groups: []instanceGroup{{created: now, ready: 2, desired: 2}},
			want:   true,
		},
		"not running is never converged": {
			state:  gql.ApplicationStateNotRunning,
			groups: []instanceGroup{{created: now, ready: 2, desired: 2}},
			want:   false,
		},
		"unknown state is never converged": {
			state:  gql.ApplicationStateUnknown,
			groups: []instanceGroup{{created: now, ready: 2, desired: 2}},
			want:   false,
		},
		"rollout in progress (two groups) is not converged": {
			state:  gql.ApplicationStateRunning,
			groups: []instanceGroup{{created: now, ready: 1, desired: 2}, {created: now, ready: 2, desired: 2}},
			want:   false,
		},
		"not all instances ready is not converged": {
			state:  gql.ApplicationStateRunning,
			groups: []instanceGroup{{created: now, ready: 1, desired: 2}},
			want:   false,
		},
		"no groups is not converged": {
			state:  gql.ApplicationStateRunning,
			groups: nil,
			want:   false,
		},
		"zero desired instances is not converged": {
			state:  gql.ApplicationStateRunning,
			groups: []instanceGroup{{created: now, ready: 0, desired: 0}},
			want:   false,
		},
	} {
		a := applicationResource{}
		t.Run(name, func(t *testing.T) {
			if got := a.isConverged(tc.state, tc.groups); got != tc.want {
				t.Errorf("isConverged = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRolloutObserved(t *testing.T) {
	since := time.Now()

	for name, tc := range map[string]struct {
		groups []instanceGroup
		want   bool
	}{
		"new group created after apply is a rollout": {
			groups: []instanceGroup{{created: since.Add(time.Second), ready: 1, desired: 2}},
			want:   true,
		},
		"group created exactly at apply time is a rollout": {
			groups: []instanceGroup{{created: since, ready: 1, desired: 2}},
			want:   true,
		},
		"multiple groups means a rollout is in progress": {
			groups: []instanceGroup{{created: since.Add(-time.Hour)}, {created: since.Add(-time.Minute)}},
			want:   true,
		},
		"single old group is not a rollout": {
			groups: []instanceGroup{{created: since.Add(-time.Hour), ready: 2, desired: 2}},
			want:   false,
		},
		"no groups is not a rollout": {
			groups: nil,
			want:   false,
		},
	} {
		a := applicationResource{}
		t.Run(name, func(t *testing.T) {
			if got := a.rolloutObserved(tc.groups, since); got != tc.want {
				t.Errorf("rolloutObserved = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestInstanceReason(t *testing.T) {
	for name, tc := range map[string]struct {
		ins  instance
		want string
	}{
		"message wins over exit reason": {
			ins:  instance{message: "ImagePullBackOff", exitReason: "Error"},
			want: "ImagePullBackOff",
		},
		"exit reason with code": {
			ins:  instance{exitReason: "OOMKilled", exitCode: new(137)},
			want: "OOMKilled (exit code 137)",
		},
		"exit reason without code": {
			ins:  instance{exitReason: "OOMKilled"},
			want: "OOMKilled",
		},
		"only exit code": {
			ins:  instance{exitCode: new(1)},
			want: "exit code 1",
		},
		"no reason": {
			ins:  instance{},
			want: "",
		},
	} {
		a := applicationResource{}
		t.Run(name, func(t *testing.T) {
			if got := a.instanceReason(tc.ins); got != tc.want {
				t.Errorf("instanceReason = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestInstanceDetail(t *testing.T) {
	ins := instance{
		name:     "testapp-abc",
		state:    gql.ApplicationInstanceStateFailing,
		restarts: 5,
		message:  `Back-off pulling image "testapp:doesnotexist"`,
	}
	a := applicationResource{}
	if got, want := a.instanceDetail(ins), `testapp-abc (state FAILING, restarts 5): Back-off pulling image "testapp:doesnotexist"`; got != want {
		t.Errorf("instanceDetail = %q, want %q", got, want)
	}

	// No restarts and no reason: just name and state.
	if got, want := a.instanceDetail(instance{name: "testapp-abc", state: gql.ApplicationInstanceStateStarting}), "testapp-abc (state STARTING)"; got != want {
		t.Errorf("instanceDetail = %q, want %q", got, want)
	}
}

func TestWaitError(t *testing.T) {
	groups := []instanceGroup{
		{
			ready:   1,
			desired: 2,
			instances: []instance{
				{name: "testapp-old", state: gql.ApplicationInstanceStateRunning, ready: true},
				{name: "testapp-new", state: gql.ApplicationInstanceStateFailing, ready: false, restarts: 3, message: "CrashLoopBackOff"},
			},
		},
	}

	a := applicationResource{}
	err := a.waitError("testapp", groups, context.DeadlineExceeded)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected error to wrap context.DeadlineExceeded, got %v", err)
	}
	for _, want := range []string{
		"timed out waiting for Application/testapp to become ready",
		"unhealthy instances:",
		"testapp-new (state FAILING, restarts 3): CrashLoopBackOff",
	} {
		mustErrorContains(t, err, want)
	}
	// Ready instances are not listed.
	if got := err.Error(); strings.Contains(got, "testapp-old") {
		t.Errorf("did not expect ready instance in error, got %q", got)
	}
}

func TestWaitError_Canceled(t *testing.T) {
	a := applicationResource{}
	err := a.waitError("testapp", nil, context.Canceled)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected error to wrap context.Canceled, got %v", err)
	}
	mustErrorContains(t, err, "canceled while waiting for Application/testapp to become ready")
	if got := err.Error(); strings.Contains(got, "timed out") {
		t.Errorf("did not expect timeout phrasing for a cancellation, got %q", got)
	}
}

func TestWaitError_NoInstanceInfo(t *testing.T) {
	a := applicationResource{}
	err := a.waitError("testapp", nil, context.DeadlineExceeded)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected error to wrap context.DeadlineExceeded, got %v", err)
	}
	mustErrorContains(t, err, "timed out waiting for Application/testapp to become ready")
	if got := err.Error(); strings.Contains(got, "unhealthy instances") {
		t.Errorf("did not expect unhealthy instances section, got %q", got)
	}
}
