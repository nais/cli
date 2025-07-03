package ui_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nais/cli/internal/option"
	"github.com/nais/cli/internal/postgres/migrate/ui"
)

type fakeTextInput struct {
	text string
}

func (f *fakeTextInput) Show(_ ...string) (string, error) {
	return f.text, nil
}

type fakeTextSelector struct {
	t        *testing.T
	selected string
	options  []string
}

func (f *fakeTextSelector) Show(_ ...string) (string, error) {
	return f.selected, nil
}

func (f *fakeTextSelector) WithOptions(options []string) ui.Selector {
	if !slices.ContainsFunc(options, func(e string) bool { return strings.Contains(e, f.selected) }) {
		f.t.Helper()
		f.t.Fatalf("selected value not in options, got %q, options: %#v", f.selected, options)
	}
	f.options = options
	return f
}

func TestUIAskForDiskSize(t *testing.T) {
	tests := map[string]struct {
		source       option.Option[int]
		enteredValue string
		expected     option.Option[int]
	}{
		"source has value and user presses Enter": {
			source:       option.Some(100),
			enteredValue: "",
			expected:     option.None[int](),
		},
		"source has value and user types in 200": {
			source:       option.Some(100),
			enteredValue: "200",
			expected:     option.Some(200),
		},
		"source has no value and user presses Enter": {
			source:       option.None[int](),
			enteredValue: "",
			expected:     option.None[int](),
		},
		"source has no value and user types in 200": {
			source:       option.None[int](),
			enteredValue: "200",
			expected:     option.Some(200),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ui.TextInput = &fakeTextInput{text: test.enteredValue}
			result := ui.AskForDiskSize(test.source)()
			if result != test.expected {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestUIAskForDiskAutoresize(t *testing.T) {
	tests := map[string]struct {
		source        option.Option[bool]
		selectedValue string
		expected      option.Option[bool]
	}{
		"source true and user presses Enter": {
			source:        option.Some(true),
			selectedValue: "Same as source (true)",
			expected:      option.Some(true),
		},
		"source true and user selects false": {
			source:        option.Some(true),
			selectedValue: "false",
			expected:      option.Some(false),
		},
		"source false and user presses Enter": {
			source:        option.Some(false),
			selectedValue: "Same as source (false)",
			expected:      option.Some(false),
		},
		"source false and user selects true": {
			source:        option.Some(false),
			selectedValue: "true",
			expected:      option.Some(true),
		},
		"source unset and user presses Enter": {
			source:        option.None[bool](),
			selectedValue: "Same as source (false)",
			expected:      option.Some(false),
		},
		"source unset and user selects true": {
			source:        option.None[bool](),
			selectedValue: "true",
			expected:      option.Some(true),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ui.TextSelector = &fakeTextSelector{t: t, selected: test.selectedValue}
			result := ui.AskForDiskAutoresize(test.source)()
			if result != test.expected {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestUIAskForTier_when_source_has_a_value_and(t *testing.T) {
	tests := map[string]struct {
		selectedValue string
		expected      option.Option[string]
	}{
		"user presses Enter": {
			selectedValue: "Same as source (db-f1-micro)",
			expected:      option.None[string](),
		},
		"user selects db-custom-2-5120": {
			selectedValue: "db-custom-2-5120",
			expected:      option.Some("db-custom-2-5120"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ui.TextSelector = &fakeTextSelector{t: t, selected: test.selectedValue}
			result := ui.AskForTier("db-f1-micro")()
			if result != test.expected {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestUIAskForTier_user_selects_Other_and_enters_a_value_it_returns_the_entered_value(t *testing.T) {
	ui.TextSelector = &fakeTextSelector{t: t, selected: "Other"}
	ui.TextInput = &fakeTextInput{text: "db-custom-16-8192"}
	result := ui.AskForTier("db-f1-micro")()
	expected := option.Some("db-custom-16-8192")

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestUIAskForTier_source_value_is_in_preset_list_of_options_it_is_only_listed_once(t *testing.T) {
	f := &fakeTextSelector{t: t, selected: "db-custom-2-5120"}
	ui.TextSelector = f
	ui.AskForTier("db-custom-2-5120")()
	if !slices.Contains(f.options, "Same as source (db-custom-2-5120)") {
		t.Errorf("expected options to contain 'Same as source (db-custom-2-5120)', got %v", f.options)
	}
	if slices.Contains(f.options, "db-custom-2-5120") {
		t.Errorf("expected options to not contain 'db-custom-2-5120', got %v", f.options)
	}
}

func TestUIAskForType(t *testing.T) {
	tests := map[string]struct {
		source        string
		selectedValue string
		expected      option.Option[string]
	}{
		"same as source": {
			source:        "POSTGRES_13",
			selectedValue: "Same as source (POSTGRES_13)",
			expected:      option.None[string](),
		},
		"selects POSTGRES_14": {
			source:        "POSTGRES_13",
			selectedValue: "POSTGRES_14",
			expected:      option.Some("POSTGRES_14"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ui.TextSelector = &fakeTextSelector{t: t, selected: test.selectedValue}
			result := ui.AskForType(test.source)()
			if result != test.expected {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestUIAskForType_source_is_POSTGRES_14_only_list_newer_versions(t *testing.T) {
	f := &fakeTextSelector{selected: "POSTGRES_15"}
	ui.TextSelector = f
	ui.AskForType("POSTGRES_14")()

	expected := []string{
		"Same as source (POSTGRES_14)",
		"POSTGRES_17",
		"POSTGRES_16",
		"POSTGRES_15",
	}
	if diff := cmp.Diff(f.options, expected); diff != "" {
		t.Errorf("options mismatch (-got +want):\n%s", diff)
	}
}
