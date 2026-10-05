package command

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/nais/cli/internal/alpha/postgres"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

func TestBranchStatusShowsRequestedAndObservedSeparately(t *testing.T) {
	for _, tc := range []struct {
		name, desired, active string
		pending               bool
	}{
		{name: "activation pending", desired: "restore", active: "main", pending: true},
		{name: "activation observed", desired: "restore", active: "restore"},
		{name: "no requested branch", active: "main"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			level := naistrix.OutputVerbosityLevelNormal
			printBranchStatus(naistrix.NewOutputWriter(&buf, &level), postgres.BranchStatus{
				DesiredActive: tc.desired, Active: tc.active,
				Branches: []postgres.Branch{{Name: "restore", State: gql.PostgresBranchStateProgressing}},
			})
			text := buf.String()
			requested, observed := tc.desired, tc.active
			if requested == "" {
				requested = "(none)"
			}
			if observed == "" {
				observed = "(none)"
			}
			if !strings.Contains(text, "Requested active branch: "+requested+"\n") || !strings.Contains(text, "Observed active branch: "+observed+"\n") || !strings.Contains(text, "restore\tProgressing") {
				t.Errorf("incomplete status: %q", text)
			}
			if got := strings.Contains(text, "Activation pending:"); got != tc.pending {
				t.Errorf("pending indicator = %v, want %v: %q", got, tc.pending, text)
			}
		})
	}
}

func TestBranchListShowsObservedAndPendingActivation(t *testing.T) {
	status := postgres.BranchStatus{
		Active: "main", DesiredActive: "restore",
		Branches: []postgres.Branch{
			{Name: "main", State: gql.PostgresBranchStateAvailable},
			{Name: "restore", State: gql.PostgresBranchStateAvailable},
		},
	}
	want := []branchListRow{
		{Name: "main", State: postgres.State(gql.PostgresBranchStateAvailable), Active: "Yes"},
		{Name: "restore", State: postgres.State(gql.PostgresBranchStateAvailable), Requested: "Yes"},
	}
	if got := branchListRows(status); !reflect.DeepEqual(got, want) {
		t.Errorf("branchListRows() = %#v, want %#v", got, want)
	}
	status.Active = "restore"
	want[0].Active = ""
	want[1].Active, want[1].Requested = "Yes", ""
	if got := branchListRows(status); !reflect.DeepEqual(got, want) {
		t.Errorf("branchListRows() after activation = %#v, want %#v", got, want)
	}
}

func TestParseRecoveryTime(t *testing.T) {
	for _, value := range []string{"", "2026-01-01T12:00:00+01:00", "2026-01-01T12:00:00", "2026-01-01T12:00:00.5Z", "9999-01-01T00:00:00Z"} {
		if _, err := parseRecoveryTime(value); err == nil {
			t.Errorf("parseRecoveryTime(%q) accepted invalid recovery time", value)
		}
	}
	got, err := parseRecoveryTime("2026-01-01T12:00:00Z")
	if err != nil || got.Location() != time.UTC {
		t.Errorf("parseRecoveryTime() = %v, %v; want UTC time", got, err)
	}
}
