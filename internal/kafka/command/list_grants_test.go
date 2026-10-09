package command

import (
	"slices"
	"testing"

	"github.com/nais/cli/internal/kafka/command/flag"
)

func TestRevokeGrantArgumentChoices(t *testing.T) {
	cmd := revokeGrant(&flag.Kafka{})
	if got, want := cmd.Args[2].Choices, []string{"read", "write", "readwrite"}; !slices.Equal(got, want) {
		t.Errorf("access choices = %v, want %v", got, want)
	}
	if !cmd.Args[2].ChoicesCaseInsensitive {
		t.Error("access choices must be case-insensitive")
	}
	for _, choice := range cmd.Args[2].Choices {
		if err := flag.KafkaTopicGrantAccess(choice).Validate(); err != nil {
			t.Errorf("access choice %q is invalid: %v", choice, err)
		}
	}
	for _, arg := range cmd.Args[:2] {
		if len(arg.Choices) != 0 {
			t.Errorf("%s must not have static choices", arg.Name)
		}
	}
}

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
