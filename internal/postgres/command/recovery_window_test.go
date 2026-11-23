package command

import "testing"

func TestRecoveryWindowDays(t *testing.T) {
	for _, tt := range []struct {
		days    int
		want    *int
		wantErr bool
	}{
		{days: 0},
		{days: 1, wantErr: true},
		{days: -2, wantErr: true},
		{days: 2, want: new(2)},
		{days: 14, want: new(14)},
		{days: 30, want: new(30)},
		{days: 31, wantErr: true},
	} {
		got, err := recoveryWindowDays(tt.days)
		if (err != nil) != tt.wantErr {
			t.Errorf("%d: error = %v, wantErr %v", tt.days, err, tt.wantErr)
			continue
		}
		if (got == nil) != (tt.want == nil) || (got != nil && *got != *tt.want) {
			t.Errorf("%d: got %v, want %v", tt.days, got, tt.want)
		}
	}
}
