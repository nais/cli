package command

import (
	"testing"

	"github.com/nais/cli/internal/kafka/command/flag"
)

func TestKafkaTopicGrantAccessValidate(t *testing.T) {
	for _, access := range []flag.KafkaTopicGrantAccess{"", "admin", "read"} {
		err := access.Validate()
		if access == "read" && err != nil {
			t.Errorf("Validate(%q) returned %v, want nil", access, err)
		}
		if access != "read" && err == nil {
			t.Errorf("Validate(%q) returned nil, want error", access)
		}
	}
}
