package command

import (
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/nais/cli/internal/workload"
	"github.com/nais/naistrix"
)

func TestParseEventQuery(t *testing.T) {
	const earlier = "2026-11-01T10:00:00Z"
	const later = "2026-11-01T11:00:00Z"
	for _, tt := range []struct {
		name            string
		limit           int
		from, to, after string
		wantErr         string
	}{
		{name: "minimum limit", limit: 1, to: later},
		{name: "maximum limit", limit: 100, from: earlier, to: later},
		{name: "below minimum", limit: 0, wantErr: "--limit must be between 1 and 100"},
		{name: "above maximum", limit: 101, wantErr: "--limit must be between 1 and 100"},
		{name: "malformed from", limit: 20, from: "yesterday", to: later, wantErr: "invalid --from timestamp"},
		{name: "malformed to", limit: 20, to: "today", wantErr: "invalid --to timestamp"},
		{name: "equal bounds", limit: 20, from: later, to: later, wantErr: "--from must be earlier than --to"},
		{name: "reversed bounds", limit: 20, from: later, to: earlier, wantErr: "--from must be earlier than --to"},
		{name: "cursor without upper bound", limit: 20, after: "opaque", wantErr: "--after requires --to"},
		{name: "fractional continuation", limit: 20, from: "2026-11-01T11:00:00.123456789+01:00", to: "2026-11-01T10:00:00.123456790Z", after: "opaque"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEventQuery(tt.limit, tt.from, tt.to, tt.after)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.Limit != tt.limit {
				t.Errorf("limit = %d, want %d", got.Limit, tt.limit)
			}
			wantTo, err := time.Parse(time.RFC3339Nano, tt.to)
			if err != nil {
				t.Fatal(err)
			}
			if got.To != wantTo.UTC() {
				t.Errorf("to = %s, want %s", got.To, wantTo.UTC())
			}
			if tt.from != "" {
				wantFrom, err := time.Parse(time.RFC3339Nano, tt.from)
				if err != nil {
					t.Fatal(err)
				}
				if got.From == nil || *got.From != wantFrom.UTC() {
					t.Errorf("from = %v, want %s", got.From, wantFrom.UTC())
				}
			} else if got.From != nil {
				t.Errorf("from = %v, want nil", got.From)
			}
			if tt.after != "" {
				if got.After == nil || *got.After != tt.after {
					t.Errorf("after = %v, want %q", got.After, tt.after)
				}
			} else if got.After != nil {
				t.Errorf("after = %v, want nil", got.After)
			}
		})
	}
}

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
	from := time.Date(2026, 11, 1, 0, 0, 0, 123456789, time.UTC)
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
		t.Fatalf("shell arguments = %q, want %q", got, want)
	}
	args := strings.Split(string(got), "\x00")
	limit, err := strconv.Atoi(args[5])
	if err != nil {
		t.Fatal(err)
	}
	query, err := ParseEventQuery(limit, args[7], args[3], args[1])
	if err != nil {
		t.Fatal(err)
	}
	if query.Limit != 20 || query.After == nil || *query.After != cursor || query.From == nil || *query.From != from || query.To != to {
		t.Errorf("next page query did not preserve the cursor and bounds: %+v", query)
	}
}
