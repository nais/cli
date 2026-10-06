package command

import (
	"testing"
	"time"
)

func TestRelativeRecoveryTime(t *testing.T) {
	now := time.Date(2026, time.October, 6, 12, 0, 0, 500_000_000, time.FixedZone("UTC+2", 2*60*60))
	for _, tt := range []struct {
		ago, want string
	}{
		{ago: "2h", want: "2026-10-06T08:00:00Z"},
		{ago: "30m", want: "2026-10-06T09:30:00Z"},
		{ago: "1h30m", want: "2026-10-06T08:30:00Z"},
		{ago: "12h", want: "2026-10-05T22:00:00Z"},
		{ago: "1.25s", want: "2026-10-06T09:59:59Z"},
	} {
		t.Run(tt.ago, func(t *testing.T) {
			got, err := resolveRecoveryTime("", tt.ago, now)
			if err != nil {
				t.Fatal(err)
			}
			if got.Format(time.RFC3339) != tt.want || got.Location() != time.UTC || got.Nanosecond() != 0 {
				t.Errorf("recovery time = %v; want %s in UTC with whole-second precision", got, tt.want)
			}
		})
	}
}

func TestRelativeRecoveryTimeRejectsInvalidDurations(t *testing.T) {
	for _, ago := range []string{"0s", "-2h", "2", "yesterday", "2d", "9999999999999999h"} {
		t.Run(ago, func(t *testing.T) {
			if _, err := resolveRecoveryTime("", ago, time.Now()); err == nil {
				t.Errorf("accepted invalid duration %q", ago)
			}
		})
	}
}
