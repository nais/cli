package command

import (
	"fmt"
	"strings"
	"time"

	"github.com/nais/cli/internal/workload"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

const DefaultEventLimit = 20

func ParseEventQuery(limit int, fromFlag, toFlag, afterFlag string) (workload.EventQuery, error) {
	query := workload.EventQuery{Limit: limit, To: time.Now().UTC()}
	if limit < 1 || limit > 100 {
		return query, fmt.Errorf("--limit must be between 1 and 100")
	}
	if toFlag != "" {
		to, err := time.Parse(time.RFC3339, toFlag)
		if err != nil {
			return query, fmt.Errorf("invalid --to timestamp: %w", err)
		}
		query.To = to.UTC()
	}
	if fromFlag != "" {
		from, err := time.Parse(time.RFC3339, fromFlag)
		if err != nil {
			return query, fmt.Errorf("invalid --from timestamp: %w", err)
		}
		if !from.Before(query.To) {
			return query, fmt.Errorf("--from must be earlier than --to")
		}
		query.From = new(from.UTC())
	}
	if afterFlag != "" {
		if toFlag == "" {
			return query, fmt.Errorf("--after requires --to from the previous page")
		}
		query.After = &afterFlag
	}
	return query, nil
}

func RenderEvents(out *naistrix.OutputWriter, page *workload.EventsPage, query workload.EventQuery, json bool) error {
	if json {
		return out.JSON(output.JSONWithPrettyOutput()).Render(page)
	}
	if len(page.Events) == 0 {
		if query.After != nil {
			out.Println("No more events found after the cursor.")
		} else {
			out.Println("No retained events in this range. The API retains workload events for 30 days.")
		}
	}
	type row struct {
		LastSeen string `heading:"Last seen"`
		Kind     string
		Instance string
		Message  string
	}
	rows := make([]row, 0, len(page.Events))
	for _, event := range page.Events {
		r := row{LastSeen: event.TriggeredAt.Format(time.RFC3339), Message: event.Message}
		switch event.Type {
		case "warning":
			r.Kind = event.Kind
			if event.Instance != nil {
				r.Instance = *event.Instance
			}
			if event.Count > 1 && event.FirstSeenAt != nil {
				r.Message += fmt.Sprintf(" (%d occurrences; first seen %s)", event.Count, event.FirstSeenAt.Format(time.RFC3339))
			}
		case "restart":
			r.Kind = "Restart"
		case "rescale":
			r.Kind = "Rescale"
		}
		rows = append(rows, r)
	}
	if len(rows) != 0 {
		if err := out.Table().Render(rows); err != nil {
			return err
		}
	}
	if page.NextAfter != nil {
		var flags strings.Builder
		// Single-quote the opaque cursor for POSIX shells, including embedded apostrophes.
		cursor := strings.ReplaceAll(*page.NextAfter, "'", "'\"'\"'")
		fmt.Fprintf(&flags, "--after '%s' --to %s --limit %d", cursor, page.To.Format(time.RFC3339Nano), query.Limit)
		if page.From != nil {
			fmt.Fprintf(&flags, " --from %s", page.From.Format(time.RFC3339Nano))
		}
		out.Printf("Next page: rerun with %s\n", flags.String())
	}
	return nil
}
