package validate

import "testing"

func TestNaisConfig(t *testing.T) {
	err := NaisConfig([]string{"testdata/nais-valid.yaml"})
	if err != nil {
		t.Errorf("NaisConfig() error = %v", err)
	}

	err = NaisConfig([]string{"testdata/nais-invalid.yaml"})
	if err == nil {
		t.Errorf("NaisConfig() error = %v", err)
	}
}
