package command

import (
	"strings"
	"testing"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/workload"
	"github.com/nais/naistrix"
	"github.com/pterm/pterm"
)

func TestRenderStatusShowsInstanceReadinessAndLastExit(t *testing.T) {
	running := app.InstanceState(gql.ApplicationInstanceStateRunning)
	pending := app.InstanceState(gql.ApplicationInstanceStatePending)
	status := &app.InstanceGroupStatus{
		Application: "web",
		Environment: "dev",
		Groups: []app.InstanceGroupInfo{{
			Image: "web:latest", ReadyInstances: 2, DesiredInstances: 4,
			Instances: []app.InstanceInfo{
				{Name: "unready", State: running, Message: "Readiness probe failed", LastExit: workload.LastExit(new("Error"), new(1), nil)},
				{Name: "pending", State: pending, Message: "No nodes available"},
				{Name: "recovered", State: running, Ready: true, Restarts: 1, Message: "Ready again", LastExit: workload.LastExit(new("OOMKilled"), new(137), nil)},
				{Name: "healthy", State: running, Ready: true, LastExit: workload.LastExit(new("Completed"), new(0), nil)},
			},
		}},
	}

	var stdout strings.Builder
	out := naistrix.NewOutputWriter(&stdout, new(naistrix.Count))
	if err := renderStatus(out, status); err != nil {
		t.Fatal(err)
	}
	got := pterm.RemoveColorFromString(stdout.String())
	words := strings.Join(strings.Fields(strings.ReplaceAll(got, "│", " ")), " ")
	for _, want := range []string{
		"Instances: 2/4 ready",
		"unready Running, not ready 0",
		"pending Waiting for placement 0",
		"recovered Running 1",
		"unready: Readiness probe failed",
		"unready last exit: Error, code 1",
		"pending: No nodes available",
		"recovered last exit: OOMKilled, code 137",
	} {
		if !strings.Contains(words, want) {
			t.Errorf("missing %q in output:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{"Ready again", "healthy last exit:", "pending last exit:"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("unexpected %q in output:\n%s", unwanted, got)
		}
	}
}
