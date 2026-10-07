package command

import (
	"testing"
	"time"
)

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
