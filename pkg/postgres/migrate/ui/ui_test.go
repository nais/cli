package ui_test

import (
	"github.com/nais/cli/pkg/option"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nais/cli/pkg/postgres/migrate/ui"
)

type fakeTextInput struct {
	text string
}

func (f *fakeTextInput) Show(_ ...string) (string, error) {
	return f.text, nil
}

type fakeTextSelector struct {
	selected string
}

func (f *fakeTextSelector) Show(_ ...string) (string, error) {
	return f.selected, nil
}

func (f *fakeTextSelector) WithOptions(options []string) ui.Selector {
	Expect(options).To(ContainElement(ContainSubstring(f.selected)))
	return f
}

var _ = Describe("Ui", func() {
	Context("AskForDiskSize", func() {
		DescribeTable("when source has", func(source option.Option[int], enteredValue string, expected option.Option[int]) {
			ui.TextInput = &fakeTextInput{text: enteredValue}
			result := ui.AskForDiskSize(source)()
			Expect(result).To(Equal(expected))
		},
			Entry("a value and user presses Enter", option.Some(100), "", option.None[int]()),
			Entry("a value and user types in 200", option.Some(100), "200", option.Some(200)),
			Entry("no value and user presses Enter", option.None[int](), "", option.None[int]()),
			Entry("no value and user types in 200", option.None[int](), "200", option.Some(200)),
		)
	})

	Context("AskForDiskAutoresize", func() {
		DescribeTable("when source has", func(source option.Option[bool], selectedValue string, expected option.Option[bool]) {
			ui.TextSelector = &fakeTextSelector{selected: selectedValue}
			result := ui.AskForDiskAutoresize(source)()
			Expect(result).To(Equal(expected))
		},
			Entry("true and user presses Enter", option.Some(true), "Same as source (true)", option.Some(true)),
			Entry("true and user selects false", option.Some(true), "false", option.Some(false)),
			Entry("false and user presses Enter", option.Some(false), "Same as source (false)", option.Some(false)),
			Entry("false and user selects true", option.Some(false), "true", option.Some(true)),
			Entry("unset and user presses Enter", option.None[bool](), "Same as source (false)", option.Some(false)),
			Entry("unset and user selects true", option.None[bool](), "true", option.Some(true)),
		)
	})

	Context("AskForTier", func() {
		DescribeTable("when source has a value and", func(selectedValue string, expected option.Option[string]) {
			ui.TextSelector = &fakeTextSelector{selected: selectedValue}
			result := ui.AskForTier("db-f1-micro")()
			Expect(result).To(Equal(expected))
		},
			Entry("user presses Enter", "Same as source (db-f1-micro)", option.None[string]()),
			Entry("user selects db-custom-2-5120", "db-custom-2-5120", option.Some("db-custom-2-5120")),
		)

		When("user selects Other and enters a value", func() {
			It("returns the entered value", func() {
				ui.TextSelector = &fakeTextSelector{selected: "Other"}
				ui.TextInput = &fakeTextInput{text: "db-custom-16-8192"}
				result := ui.AskForTier("db-f1-micro")()
				Expect(result).To(Equal(option.Some("db-custom-16-8192")))
			})
		})
	})

	Context("AskForType", func() {
		DescribeTable("", func(source string, selectedValue string, expected option.Option[string]) {
			ui.TextSelector = &fakeTextSelector{selected: selectedValue}
			result := ui.AskForType(source)()
			Expect(result).To(Equal(expected))
		},
			EntryDescription("source: %s, selected: %s => %v"),
			Entry(nil, "POSTGRES_13", "Same as source (POSTGRES_13)", option.None[string]()),
			Entry(nil, "POSTGRES_13", "POSTGRES_14", option.Some("POSTGRES_14")),
		)
	})
})
