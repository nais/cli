package ui

import (
	"fmt"
	"github.com/nais/cli/pkg/option"
	"github.com/pterm/pterm"
	"log"
	"strconv"
	"strings"
)

const (
	otherOption              = "Other"
	sameAsSourceOptionPrefix = "Same as source"
)

var CmdStyle = pterm.NewStyle(pterm.FgLightMagenta)
var LinkStyle = pterm.NewStyle(pterm.FgLightBlue, pterm.Underscore)
var YamlStyle = pterm.NewStyle(pterm.FgLightYellow)

func stringCaster(s string) string { return s }
func boolCaster(s string) bool     { return s == "true" }

func askForOption[T any](prompt string, sourceValue T, options []string, caster func(string) T, otherHandler func() string) func() option.Option[T] {
	return func() option.Option[T] {
		source := fmt.Sprintf("%s (%v)", sameAsSourceOptionPrefix, sourceValue)
		options = append([]string{source}, options...)
		if otherHandler != nil {
			options = append(options, otherOption)
		}
		pterm.Println()
		selected, err := pterm.DefaultInteractiveSelect.
			WithOptions(options).
			WithMaxHeight(len(options)).
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

var tierOptions = []string{
	"db-custom-1-3840",
	"db-custom-2-5120",
	"db-custom-2-7680",
	"db-custom-4-15360",
}

func AskForTier(sourceTier string) func() option.Option[string] {
	var options []string
	for _, tier := range tierOptions {
		if tier != sourceTier {
			options = append(options, tier)
		}
	}
	return askForOption("Select a tier for the target instance", sourceTier, options, stringCaster, func() string {
		pterm.Println("Check the documentation for possible options:")
		LinkStyle.Printfln("\thttps://doc.nais.io/persistence/postgres/reference/#server-size")
		tier, err := pterm.DefaultInteractiveTextInput.Show("Enter the tier for the target instance")
		if err != nil {
			log.Fatalf("Error while creating text UI: %v", err)
			return ""
		}
		return tier
	})
}

var typeToVersion = map[string]int{
	"POSTGRES_11": 11,
	"POSTGRES_12": 12,
	"POSTGRES_13": 13,
	"POSTGRES_14": 14,
	"POSTGRES_15": 15,
	"POSTGRES_16": 16,
}

func AskForType(sourceType string) func() option.Option[string] {
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
	return askForOption("Select a type for the target instance", sourceType, options, stringCaster, nil)
}

func AskForDiskAutoresize(sourceDiskAutoresize option.Option[bool]) func() option.Option[bool] {
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

func AskForDiskSize(sourceDiskSize option.Option[int]) func() option.Option[int] {
	sourceSize := "<nais default>"
	sourceDiskSize.Do(func(v int) {
		sourceSize = fmt.Sprintf("%d GB", v)
	})
	var ask func() option.Option[int]
	ask = func() option.Option[int] {
		pterm.Println()
		pterm.Println("Disk size is in GB, and must be greater than or equal to 10.")
		msg := fmt.Sprintf("Enter the disk size for the target instance. Leave empty to use same as source (%s)", sourceSize)
		diskSize, err := pterm.DefaultInteractiveTextInput.Show(msg)
		if err != nil {
			log.Fatalf("Error while creating text UI: %v", err)
			return option.None[int]()
		}
		if diskSize == "" {
			return option.None[int]()
		}
		size, err := strconv.Atoi(diskSize)
		if err != nil {
			pterm.Error.Println("Disk size must be a number")
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
