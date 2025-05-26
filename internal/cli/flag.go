package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type count int

type flagTypes interface{ int | bool | string | count }

func setupFlag(name, short, usage string, value any, flags *pflag.FlagSet) {
	if len(short) > 1 {
		panic("short flag must be a single character")
	}

	switch ptr := value.(type) {
	case *string:
		if short == "" {
			flags.StringVar(ptr, name, "", usage)
		} else {
			flags.StringVarP(ptr, name, short, "", usage)
		}
	case *bool:
		if short == "" {
			flags.BoolVar(ptr, name, false, usage)
		} else {
			flags.BoolVarP(ptr, name, short, false, usage)
		}
	case *int:
		if short == "" {
			flags.CountVar(ptr, name, usage)
		} else {
			flags.CountVarP(ptr, name, short, usage)
		}
	default:
		panic(fmt.Sprintf("unknown flag type: %T", value))
	}
}

type FlagOption func(*cobra.Command, string)

func FlagRequired() FlagOption {
	return func(cmd *cobra.Command, name string) {
		if err := cmd.MarkFlagRequired(name); err != nil {
			panic("failed to mark flag as required: " + err.Error())
		}
	}
}
