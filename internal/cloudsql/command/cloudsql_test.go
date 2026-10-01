package command

import (
	"bytes"
	"strings"
	"testing"
)

func TestWarnIfLegacyAlias(t *testing.T) {
	for name, tt := range map[string]struct {
		args []string
		warn bool
	}{
		"postgres":       {[]string{"postgres", "list"}, true},
		"pg":             {[]string{"pg", "psql", "app"}, true},
		"flag before":    {[]string{"--team", "x", "postgres"}, false},
		"cloudsql":       {[]string{"cloudsql", "list"}, false},
		"alpha postgres": {[]string{"alpha", "postgres", "list"}, false},
		"only flags":     {[]string{"--help"}, false},
		"no args":        {nil, false},
	} {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			warnIfLegacyAlias(tt.args, &buf)
			if got := strings.Contains(buf.String(), "has moved to nais cloudsql"); got != tt.warn {
				t.Fatalf("warned=%v, want %v (%q)", got, tt.warn, buf.String())
			}
		})
	}
}
