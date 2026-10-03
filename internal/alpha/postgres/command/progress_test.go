package command

import (
	"bytes"
	"strings"
	"testing"
)

func TestAccessSpinnerStopsBeforeNextOutput(t *testing.T) {
	var out bytes.Buffer
	spinner := newAccessSpinner(&out)
	spinner.Stop()
	if got := out.String(); !strings.Contains(got, "Preparing personal Postgres access...") {
		t.Fatalf("spinner did not show progress: %q", got)
	}
	out.WriteString("Access ready.\n")
	if got := out.String(); !strings.HasSuffix(got, "\r\x1b[2KAccess ready.\n") {
		t.Errorf("spinner did not clear its line before the next output: %q", got)
	}
}
