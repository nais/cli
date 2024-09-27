package kubeconfigcmd

// Generate data driven test for getTenantFromEmail

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTenantFromEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "valid NAV email",
			email:    "kyrre.havik@nav.no",
			expected: "nav",
		},
		{
			name:     "valid SSB email",
			email:    "kyrre.havik@ssb.no",
			expected: "ssb",
		},
		{
			name:     "invalid email",
			email:    "kyrre.havik",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenant, err := getTenantFromEmail(tt.email)
			if tt.expected == "" {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, tenant)
		})
	}
}
