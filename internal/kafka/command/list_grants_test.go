package command

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/kafka"
	"github.com/nais/cli/internal/kafka/command/flag"
)

func TestListGrantsCommand(t *testing.T) {
	command := listGrants(&flag.Kafka{GlobalFlags: &flags.GlobalFlags{}})

	if command.Name != "list-grants" {
		t.Errorf("Name = %q, want %q", command.Name, "list-grants")
	}
	if len(command.Args) != 1 || command.Args[0].Name != "topic" {
		t.Errorf("Args = %#v, want a single topic argument", command.Args)
	}
	if command.ValidateFunc == nil {
		t.Error("ValidateFunc is nil")
	}
	if command.AutoCompleteFunc == nil {
		t.Error("AutoCompleteFunc is nil")
	}
}

func TestKafkaTopicGrantAccessAutoComplete(t *testing.T) {
	access := flag.KafkaTopicGrantAccess("")
	completions, description := access.AutoComplete(t.Context(), nil, "", nil)

	if want := []string{"read", "write", "readwrite"}; !reflect.DeepEqual(completions, want) {
		t.Errorf("completions = %q, want %q", completions, want)
	}
	if description != "Available access levels." {
		t.Errorf("description = %q, want %q", description, "Available access levels.")
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

func TestRevokeGrantCommand(t *testing.T) {
	command := revokeGrant(&flag.Kafka{GlobalFlags: &flags.GlobalFlags{}})

	if command.Name != "revoke-grant" {
		t.Errorf("Name = %q, want %q", command.Name, "revoke-grant")
	}
	if len(command.Args) != 3 || command.Args[0].Name != "topic" || command.Args[1].Name != "username" || command.Args[2].Name != "access" {
		t.Errorf("Args = %#v, want topic, username, and access arguments", command.Args)
	}
	if command.ValidateFunc == nil {
		t.Error("ValidateFunc is nil")
	}
	if command.AutoCompleteFunc == nil {
		t.Error("AutoCompleteFunc is nil")
	}
}

func TestGrantTableHeadings(t *testing.T) {
	var buf bytes.Buffer
	out := newTestOutputWriter(&buf)

	if err := out.Table().Render([]kafka.Grant{{
		WorkloadName: "consumer",
		TeamName:     "other-team",
		Access:       "READ",
	}}); err != nil {
		t.Fatal(err)
	}

	for _, heading := range []string{"Subject", "Team", "Access level"} {
		if !strings.Contains(buf.String(), heading) {
			t.Errorf("table output missing heading %q:\n%s", heading, buf.String())
		}
	}

	grantType := reflect.TypeFor[kafka.Grant]()
	for _, field := range []struct {
		name    string
		heading string
	}{
		{name: "WorkloadName", heading: "Subject"},
		{name: "TeamName", heading: "Team"},
		{name: "Access", heading: "Access level"},
	} {
		grantField, ok := grantType.FieldByName(field.name)
		if !ok {
			t.Fatalf("field %q not found", field.name)
		}
		got := grantField.Tag.Get("heading")
		if got != field.heading {
			t.Errorf("%s heading = %q, want %q", field.name, got, field.heading)
		}
	}
}
