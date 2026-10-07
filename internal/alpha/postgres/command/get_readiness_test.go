package command

import (
	"testing"

	"github.com/nais/cli/internal/alpha/postgres"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi/gql"
)

func TestPostgresReadiness(t *testing.T) {
	available := gql.PostgresBranchStateAvailable
	progressing := gql.PostgresBranchStateProgressing
	degraded := gql.PostgresBranchStateDegraded
	for _, tt := range []struct {
		name, active, desired, wantStatus string
		branches                          []postgres.Branch
	}{
		{name: "available active branch", active: "main", desired: "main", branches: []postgres.Branch{{Name: "main", State: available}}, wantStatus: "Ready"},
		{name: "inactive degraded branch does not degrade active summary", active: "main", desired: "main", branches: []postgres.Branch{{Name: "main", State: available}, {Name: "failed-restore", State: degraded}}, wantStatus: "Ready"},
		{name: "progressing active branch", active: "main", branches: []postgres.Branch{{Name: "main", State: progressing}}, wantStatus: "Not ready yet"},
		{name: "degraded active branch", active: "main", branches: []postgres.Branch{{Name: "main", State: degraded}}, wantStatus: "Needs attention"},
		{name: "available requested branch still awaits observed activation", active: "main", desired: "restored", branches: []postgres.Branch{{Name: "main", State: available}, {Name: "restored", State: available}}, wantStatus: "Updating"},
		{name: "progressing requested branch", active: "main", desired: "restored", branches: []postgres.Branch{{Name: "main", State: available}, {Name: "restored", State: progressing}}, wantStatus: "Updating"},
		{name: "degraded requested branch", active: "main", desired: "restored", branches: []postgres.Branch{{Name: "main", State: available}, {Name: "restored", State: degraded}}, wantStatus: "Needs attention"},
		{name: "pending activation does not hide active degradation", active: "main", desired: "restored", branches: []postgres.Branch{{Name: "main", State: degraded}, {Name: "restored", State: available}}, wantStatus: "Needs attention"},
		{name: "provisioning before active branch observation", desired: "main", branches: []postgres.Branch{{Name: "main", State: progressing}}, wantStatus: "Not ready yet"},
		{name: "available desired branch is not yet observed", desired: "main", branches: []postgres.Branch{{Name: "main", State: available}}, wantStatus: "Not ready yet"},
		{name: "failed initial provisioning", desired: "main", branches: []postgres.Branch{{Name: "main", State: degraded}}, wantStatus: "Needs attention"},
		{name: "requested branch not yet listed", active: "main", desired: "restored", branches: []postgres.Branch{{Name: "main", State: available}}, wantStatus: "Updating"},
		{name: "no observed or desired branch", wantStatus: "Unknown"},
		{name: "active branch missing from branch list", active: "main", branches: []postgres.Branch{{Name: "restored", State: available}}, wantStatus: "Unknown"},
		{name: "active branch missing during pending activation", active: "main", desired: "restored", branches: []postgres.Branch{{Name: "restored", State: available}}, wantStatus: "Unknown"},
		{name: "active state missing", active: "main", branches: []postgres.Branch{{Name: "main"}}, wantStatus: "Unknown"},
		{name: "unknown active state", active: "main", branches: []postgres.Branch{{Name: "main", State: "FUTURE_STATE"}}, wantStatus: "Unknown"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := postgresReadiness(postgres.BranchStatus{Active: tt.active, DesiredActive: tt.desired, Branches: tt.branches})
			if got != tt.wantStatus {
				t.Errorf("status = %q; want %q", got, tt.wantStatus)
			}
		})
	}
}

func TestPostgresStatusRemainsOnlyForBranches(t *testing.T) {
	pg := Postgres(&flags.GlobalFlags{})
	var branchStatus bool
	for _, cmd := range pg.SubCommands {
		if cmd.Name == "status" {
			t.Error("top-level status must not be registered")
		}
		for _, alias := range cmd.Aliases {
			if alias == "status" {
				t.Error("top-level status must not be aliased")
			}
		}
		if cmd.Name == "branch" {
			for _, child := range cmd.SubCommands {
				branchStatus = branchStatus || child.Name == "status"
			}
		}
	}
	if !branchStatus {
		t.Error("branch status must remain registered")
	}
}
