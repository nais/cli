package ui

import (
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"

	"github.com/nais/cli/internal/option"
	"github.com/pterm/pterm"
)

const (
	otherOption              = "Other"
	sameAsSourceOptionPrefix = "Same as source"
)

var (
	CmdStyle  = pterm.NewStyle(pterm.FgLightMagenta)
	LinkStyle = pterm.NewStyle(pterm.FgLightBlue, pterm.Underscore)
	YamlStyle = pterm.NewStyle(pterm.FgLightYellow)
)

func stringCaster(s string) string { return s }
func boolCaster(s string) bool     { return s == "true" }

type Prompter interface {
	Show(text ...string) (string, error)
}

var TextInput Prompter = pterm.DefaultInteractiveTextInput

type Selector interface {
	Prompter
	WithOptions(options []string) Selector
}

type textSelector struct {
	defaultSelector *pterm.InteractiveSelectPrinter
}

func (t *textSelector) Show(text ...string) (string, error) {
	return t.defaultSelector.Show(text...)
}

func (t *textSelector) WithOptions(options []string) Selector {
	return &textSelector{
		defaultSelector: pterm.DefaultInteractiveSelect.
			WithOptions(options).
			WithMaxHeight(len(options)),
	}
}

var TextSelector Selector = &textSelector{defaultSelector: &pterm.DefaultInteractiveSelect}

// askForOption is a generic function to ask for an option from a list of options.
//
// It returns a function that can be called to ask for the option.
// The function returns the selected option as an Option[T].
// If the selected option is the "Same as source" option, it returns None[T].
// If the selected option is "Other", it calls the otherHandler function to ask for the value.
// The selected value is then cast to the desired type T using caster function, and returned as Some[T].
func askForOption[T any](prompt string, sourceValue T, options []string, caster func(string) T, otherHandler func() string) func() option.Option[T] {
	return func() option.Option[T] {
		source := fmt.Sprintf("%s (%v)", sameAsSourceOptionPrefix, sourceValue)
		options = append([]string{source}, options...)
		if otherHandler != nil {
			options = append(options, otherOption)
		}
		pterm.Println()
		selected, err := TextSelector.
			WithOptions(options).
			Show(prompt)
		if err != nil {
			log.Fatalf("Error while creating text UI: %v", err)
			return option.None[T]()
		}
		if selected == otherOption {
			selected = otherHandler()
		}
		if strings.HasPrefix(selected, sameAsSourceOptionPrefix) {
			return option.None[T]()
		}
		return option.Some(caster(selected))
	}
}

// Suggested options for tier when asking user for a target tier.
var tierOptions = []string{
	"db-custom-1-3840",
	"db-custom-2-5120",
	"db-custom-2-7680",
	"db-custom-4-15360",
}

var AskForTier = askForTier

// askForTier asks for a tier for the target instance.
//
// It returns a function that can be called to ask for the tier.
// The function returns the selected tier as an Option[string].
// If the selected tier is the "Same as source" tier, it returns None[string].
// If the selected tier is "Other", it asks the user to enter a custom tier.
// The selected value is returned as Some[string].
func askForTier(sourceTier string) func() option.Option[string] {
	var options []string
	for _, tier := range tierOptions {
		if tier != sourceTier {
			options = append(options, tier)
		}
	}
	return askForOption("Select a tier for the target instance", sourceTier, options, stringCaster, func() string {
		pterm.Println("Check the documentation for possible options:")
		LinkStyle.Printfln("\thttps://doc.nais.io/persistence/postgres/reference/#server-size")
		tier, err := TextInput.Show("Enter the tier for the target instance")
		if err != nil {
			log.Fatalf("Error while creating text UI: %v", err)
			return ""
		}
		return tier
	})
}

// Mapping from instance type to version.
var typeToVersion = map[string]int{
	"POSTGRES_11": 11,
	"POSTGRES_12": 12,
	"POSTGRES_13": 13,
	"POSTGRES_14": 14,
	"POSTGRES_15": 15,
	"POSTGRES_16": 16,
	"POSTGRES_17": 17,
}

var AskForType = askForType

// askForType asks for a type for the target instance.
//
// It returns a function that can be called to ask for the type.
// The function returns the selected type as an Option[string].
// If the selected type is the "Same as source" type, it returns None[string].
// It is not possible to select a type (postgres version) less than source.
// The selected value is returned as Some[string].
func askForType(sourceType string) func() option.Option[string] {
	sourceVersion := typeToVersion[sourceType]
	var options []string
	for k, v := range typeToVersion {
		if v > sourceVersion {
			options = append(options, k)
		}
	}
	if len(options) == 0 {
		return func() option.Option[string] { return option.None[string]() }
	}
	slices.Sort(options)
	slices.Reverse(options)
	return askForOption("Select a type for the target instance", sourceType, options, stringCaster, nil)
}

var AskForDiskAutoresize = askForDiskAutoresize

// askForDiskAutoresize asks for disk autoresize for the target instance.
//
// It returns a function that can be called to ask for disk autoresize.
// The function returns the selected disk autoresize as an Option[bool].
// If the source was unset, source is considered false (the nais default), and the "Same as source" option returns Some(false).
// It always returns Some(value), where value is the selected option.
func askForDiskAutoresize(sourceDiskAutoresize option.Option[bool]) func() option.Option[bool] {
	var options []string
	autoresize := false
	sourceDiskAutoresize.Do(func(v bool) {
		autoresize = v
	})
	if autoresize {
		options = append(options, "false")
	} else {
		options = append(options, "true")
	}
	return func() option.Option[bool] {
		targetDiskAutoresize := askForOption("Enable disk autoresize for the target instance?", autoresize, options, boolCaster, nil)()
		sourceDiskAutoresize.OrValue(false).Do(func(v bool) {
			targetDiskAutoresize = targetDiskAutoresize.OrValue(v)
		})
		return targetDiskAutoresize
	}
}

var AskForDiskSize = askForDiskSize

// askForDiskSize asks for disk size for the target instance.
//
// It returns a function that can be called to ask for the disk size.
// The function returns the selected disk size as an Option[int].
// If the user enters a blank string, it returns None[int].
// If the user enters a number, it returns Some(value), where value is the entered number.
func askForDiskSize(sourceDiskSize option.Option[int]) func() option.Option[int] {
	sourceSize := "<nais default>"
	sourceDiskSize.Do(func(v int) {
		sourceSize = fmt.Sprintf("%d GB", v)
	})
	var ask func() option.Option[int]
	ask = func() option.Option[int] {
		pterm.Println()
		pterm.Println("Disk size is in GB, and must be greater than or equal to 10.")
		msg := fmt.Sprintf("Enter the disk size for the target instance. Leave empty to use same as source (%s)", sourceSize)
		diskSize, err := TextInput.Show(msg)
		if err != nil {
			log.Fatalf("Error while creating text UI: %v", err)
			return option.None[int]()
		}
		if diskSize == "" {
			return option.None[int]()
		}
		size, err := strconv.Atoi(diskSize)
		if err != nil {
			pterm.Error.Println("Disk size must be a whole number")
			return ask()
		}
		if size < 10 {
			pterm.Error.Println("Disk size must be greater than or equal to 10")
			return ask()
		}
		return option.Some(size)
	}
	return ask
}
