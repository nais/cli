package flags

import "testing"

func TestEnvironmentIsOnPrem(t *testing.T) {
	for _, tt := range []struct {
		name   Environment
		onPrem bool
	}{
		{name: "dev-fss", onPrem: true},
		{name: "prod-fss", onPrem: true},
		{name: "dev-gcp"},
		{name: "prod-gcp"},
		{name: "staging"},
		{name: "fss-test"},
		{name: ""},
	} {
		t.Run(string(tt.name), func(t *testing.T) {
			if got := tt.name.IsOnPrem(); got != tt.onPrem {
				t.Errorf("IsOnPrem() = %v, want %v", got, tt.onPrem)
			}
		})
	}
}
