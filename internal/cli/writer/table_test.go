package writer_test

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nais/cli/internal/cli/writer"
)

func TestTable_SingleLevel(t *testing.T) {
	var buf bytes.Buffer
	table := writer.NewTable(&buf)
	table.AddColumn("First name", "Name")
	table.AddColumn("Age", "Age")

	data := []struct {
		Name string
		Age  int
	}{
		{"Alice", 30},
		{"Bob", 25},
	}

	err := table.Write(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := "\x1b[96m\x1b[96mFirst name\x1b[90m\x1b[90m | \x1b[0m\x1b[96m\x1b[0m\x1b[96mAge\x1b[0m\n\x1b[96m\x1b[0m\x1b[0mAlice     \x1b[90m\x1b[90m | \x1b[0m\x1b[0m30 \nBob       \x1b[90m\x1b[90m | \x1b[0m\x1b[0m25 \n\n"
	if diff := cmp.Diff(buf.String(), expected); diff != "" {
		t.Errorf("unexpected output (-got +want):\n%s", diff)
	}
}
