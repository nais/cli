package cli

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Count int

type AutoCompleter interface {
	AutoComplete(ctx context.Context, args []string, toComplete string, flags any) (completions []string, activeHelp string)
}

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

func setupFlags(cmd *cobra.Command, flags any, flagSet *pflag.FlagSet) {
	if flags == nil {
		return
	}

	re := regexp.MustCompile(`\|([^|]+)\|`)
	normalizeUsage := func(usage string) string {
		return re.ReplaceAllStringFunc(usage, func(s string) string {
			trimmed := strings.Trim(s, "|")
			return "`" + strings.ToUpper(trimmed) + "`"
		})
	}

	fields := reflect.TypeOf(flags).Elem()
	values := reflect.ValueOf(flags).Elem()
	for i := range fields.NumField() {
		field := fields.Field(i)
		value := values.Field(i)

		if !field.IsExported() {
			continue
		}

		if value.Kind() == reflect.Pointer && value.Elem().Kind() == reflect.Struct {
			continue
		}

		flagName, ok := field.Tag.Lookup("name")
		if !ok {
			flagName = strings.ToLower(field.Name)
		}

		flagUsage, ok := field.Tag.Lookup("usage")
		if !ok {
			flagUsage = field.Name
		}
		flagShort, _ := field.Tag.Lookup("short")

		if !value.CanAddr() {
			panic(fmt.Sprintf("field %v is not addressable, cannot set up flag", field.Name))
		}

		actualValue := value.Addr().Interface()
		setupFlag(flagName, flagShort, normalizeUsage(flagUsage), unwrap(actualValue), flagSet)

		if v, ok := actualValue.(AutoCompleter); ok {
			_ = cmd.RegisterFlagCompletionFunc(flagName,
				func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
					completions, activeHelp := v.AutoComplete(cmd.Context(), args, toComplete, flags)
					if activeHelp != "" {
						completions = cobra.AppendActiveHelp(completions, activeHelp)
					}
					return completions, cobra.ShellCompDirectiveNoFileComp
				})
		} else {
			// TODO: add metric for flags that do not support autocomplete
			_ = cmd.RegisterFlagCompletionFunc(flagName, noAutocomplete())
		}
	}
}

func unwrap(value any) any {
	v := reflect.ValueOf(value)
	switch v.Elem().Kind() {
	case reflect.String:
		var t *string
		return v.Convert(reflect.TypeOf(t)).Interface()
	default:
		return value
	}
}
