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
	options  []string
}

func (f *fakeTextSelector) Show(_ ...string) (string, error) {
	return f.selected, nil
}

func (f *fakeTextSelector) WithOptions(options []string) ui.Selector {
	f.options = options
	Expect(options).To(ContainElement(ContainSubstring(f.selected)))
	return f
}

func (f *fakeTextSelector) Options() []string {
	return f.options
}

var _ = Describe("Ui", func() {
	Context("AskForDiskSize", func() {
		When("user presses enter", func() {
			BeforeEach(func() {
				ui.TextInput = &fakeTextInput{text: ""}
			})

			It("should return the default value", func() {
				result := ui.AskForDiskSize(option.Some(100))()
				Expect(result).To(Equal(option.None[int]()))
			})
		})

		When("user types in 200", func() {
			BeforeEach(func() {
				ui.TextInput = &fakeTextInput{text: "200"}
			})

			It("should return the entered value", func() {
				result := ui.AskForDiskSize(option.Some(100))()
				Expect(result).To(Equal(option.Some(200)))
			})
		})
	})

	Context("AskForDiskAutoresize", func() {
		When("source has true", func() {
			When("user presses enter", func() {
				BeforeEach(func() {
					ui.TextSelector = &fakeTextSelector{selected: "Same as source (true)"}
				})

				It("should return true", func() {
					result := ui.AskForDiskAutoresize(option.Some(true))()
					Expect(result).To(Equal(option.Some(true)))
				})
			})

			When("user selects false", func() {
				BeforeEach(func() {
					ui.TextSelector = &fakeTextSelector{selected: "false"}
				})

				It("should return the entered value", func() {
					result := ui.AskForDiskAutoresize(option.Some(true))()
					Expect(result).To(Equal(option.Some(false)))
				})
			})
		})

		When("source is not set", func() {
			When("user presses enter", func() {
				BeforeEach(func() {
					ui.TextSelector = &fakeTextSelector{selected: "Same as source (false)"}
				})

				It("should return false", func() {
					result := ui.AskForDiskAutoresize(option.None[bool]())()
					Expect(result).To(Equal(option.Some(false)))
				})
			})

			When("user selects true", func() {
				BeforeEach(func() {
					ui.TextSelector = &fakeTextSelector{selected: "true"}
				})

				It("should return the entered value", func() {
					result := ui.AskForDiskAutoresize(option.None[bool]())()
					Expect(result).To(Equal(option.Some(true)))
				})
			})
		})

		When("source has false", func() {
			When("user presses enter", func() {
				BeforeEach(func() {
					ui.TextSelector = &fakeTextSelector{selected: "Same as source (false)"}
				})

				It("should return false", func() {
					result := ui.AskForDiskAutoresize(option.Some(false))()
					Expect(result).To(Equal(option.Some(false)))
				})
			})

			When("user selects true", func() {
				BeforeEach(func() {
					ui.TextSelector = &fakeTextSelector{selected: "true"}
				})

				It("should return the entered value", func() {
					result := ui.AskForDiskAutoresize(option.Some(false))()
					Expect(result).To(Equal(option.Some(true)))
				})
			})
		})
	})
})
