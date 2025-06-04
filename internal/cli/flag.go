package cli

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/pflag"
)

type Count int

func setupFlag(name, short, usage string, value any, flags *pflag.FlagSet) {
	if len(short) > 1 {
		panic("short flag must be a single character")
	}

	switch ptr := value.(type) {
	case *string:
		if short == "" {
			flags.StringVar(ptr, name, *ptr, usage)
		} else {
			flags.StringVarP(ptr, name, short, *ptr, usage)
		}
	case *bool:
		if short == "" {
			flags.BoolVar(ptr, name, *ptr, usage)
		} else {
			flags.BoolVarP(ptr, name, short, *ptr, usage)
		}
	case *uint:
		if short == "" {
			flags.UintVar(ptr, name, *ptr, usage)
		} else {
			flags.UintVarP(ptr, name, short, *ptr, usage)
		}
	case *[]string:
		if short == "" {
			flags.StringSliceVar(ptr, name, *ptr, usage)
		} else {
			flags.StringSliceVarP(ptr, name, short, *ptr, usage)
		}
	case *int:
		if short == "" {
			flags.IntVar(ptr, name, *ptr, usage)
		} else {
			flags.IntVarP(ptr, name, short, *ptr, usage)
		}
	case *Count:
		intPtr := (*int)(ptr)

		if short == "" {
			flags.CountVar(intPtr, name, usage)
		} else {
			flags.CountVarP(intPtr, name, short, usage)
		}
	default:
		panic(fmt.Sprintf("unknown flag type: %T", value))
	}
}

func setupFlags(flags any, flagSet *pflag.FlagSet) {
	fields := reflect.TypeOf(flags).Elem()
	values := reflect.ValueOf(flags).Elem()

	for i := range fields.NumField() {
		field := fields.Field(i)
		value := values.Field(i)

		if !field.IsExported() {
			continue
		}

		// or is it just optional?
		flagName, ok := field.Tag.Lookup("name")
		if !ok {
			flagName = strings.ToLower(field.Name)
		}

		flagUsage, ok := field.Tag.Lookup("usage")
		if !ok {
			flagUsage = field.Name
		}
		flagShort, ok := field.Tag.Lookup("short")
		if !ok {
			flagShort = ""
		}

		if !value.CanAddr() {
			panic(fmt.Sprintf("field %v is not addressable, cannot set up flag", field.Name))
		}

		setupFlag(flagName, flagShort, flagUsage, value.Addr().Interface(), flagSet)
	}
}
