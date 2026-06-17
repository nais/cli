package resource

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/nais/cli/internal/naisapi/gql"
)

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
