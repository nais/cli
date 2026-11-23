package command

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/nais/cli/internal/workload"
	"github.com/nais/naistrix"
)

func TestRenderEmptyEventsPage(t *testing.T) {
	for _, tt := range []struct {
		name  string
		after *string
		want  string
	}{
		{name: "first page", want: "No retained events in this range. The API retains workload events for 30 days.\n"},
		{name: "continuation page", after: new("previous"), want: "No more events found after the cursor.\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var stdout strings.Builder
			out := naistrix.NewOutputWriter(&stdout, new(naistrix.Count))
			if err := RenderEvents(out, &workload.EventsPage{}, workload.EventQuery{After: tt.after}, false); err != nil {
				t.Fatal(err)
			}
			if got := stdout.String(); got != tt.want {
				t.Errorf("stdout = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderEventsNextPageShellRoundTrip(t *testing.T) {
	cursor := "cursor with a space, 'quotes', a \\ backslash, $HOME, and `printf changed`\nnext line"
	from := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(24*time.Hour + 123*time.Nanosecond)
	page := &workload.EventsPage{
		Events: []workload.Event{{Type: "warning", Kind: "Failed", TriggeredAt: from}},
		From:   &from, To: to, NextAfter: &cursor,
	}
	var stdout strings.Builder
	out := naistrix.NewOutputWriter(&stdout, new(naistrix.Count))
	if err := RenderEvents(out, page, workload.EventQuery{Limit: 20}, false); err != nil {
		t.Fatal(err)
	}
	_, flags, ok := strings.Cut(stdout.String(), "Next page: rerun with ")
	if !ok {
		t.Fatalf("no next page hint in %q", stdout.String())
	}
	got, err := exec.CommandContext(t.Context(), "sh", "-c", "set -- "+flags+`printf '%s\000' "$@"`).Output()
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Join([]string{"--after", cursor, "--to", to.Format(time.RFC3339Nano), "--limit", "20", "--from", from.Format(time.RFC3339Nano), ""}, "\x00")
	if string(got) != want {
		t.Errorf("shell arguments = %q, want %q", got, want)
	}
}
