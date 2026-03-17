package cliflags

import "strings"

var defaultValueTakingFlags = []string{
	"-t",
	"--team",
	"-e",
	"--environment",
	"--config",
	"--run-name",
}

// UniqueFlagValues returns unique values for a short/long CLI flag from args.
// It supports forms: -e value, --environment value, -e=value, --environment=value.
func UniqueFlagValues(args []string, shortFlag, longFlag string) []string {
	seen := map[string]struct{}{}
	values := make([]string, 0)

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			break
		}
		switch {
		case arg == shortFlag || arg == longFlag:
			if i+1 >= len(args) {
				continue
			}
			next := args[i+1]
			if strings.HasPrefix(next, "-") || next == "" {
				continue
			}
			if _, ok := seen[next]; !ok {
				seen[next] = struct{}{}
				values = append(values, next)
			}
			i++
		case strings.HasPrefix(arg, longFlag+"="):
			val := strings.TrimPrefix(arg, longFlag+"=")
			if val != "" {
				if _, ok := seen[val]; !ok {
					seen[val] = struct{}{}
					values = append(values, val)
				}
			}
		case strings.HasPrefix(arg, shortFlag+"="):
			val := strings.TrimPrefix(arg, shortFlag+"=")
			if val != "" {
				if _, ok := seen[val]; !ok {
					seen[val] = struct{}{}
					values = append(values, val)
				}
			}
		}
	}

	return values
}

// CountFlagOccurrences counts valid value occurrences for short/long CLI flags.
func CountFlagOccurrences(args []string, shortFlag, longFlag string) int {
	count := 0

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			break
		}
		switch {
		case arg == shortFlag || arg == longFlag:
			if i+1 >= len(args) {
				continue
			}
			next := args[i+1]
			if strings.HasPrefix(next, "-") || next == "" {
				continue
			}
			count++
			i++
		case strings.HasPrefix(arg, longFlag+"="):
			if strings.TrimPrefix(arg, longFlag+"=") != "" {
				count++
			}
		case strings.HasPrefix(arg, shortFlag+"="):
			if strings.TrimPrefix(arg, shortFlag+"=") != "" {
				count++
			}
		}
	}

	return count
}

// FirstFlagValue returns the first valid value for a short/long CLI flag from args.
// It supports forms: -t value, --team value, -t=value, --team=value.
func FirstFlagValue(args []string, shortFlag, longFlag string) string {
	for i := range args {
		arg := args[i]
		if arg == "--" {
			break
		}

		if after, ok := strings.CutPrefix(arg, longFlag+"="); ok {
			if after != "" {
				return after
			}
			continue
		}
		if after, ok := strings.CutPrefix(arg, shortFlag+"="); ok {
			if after != "" {
				return after
			}
			continue
		}

		if arg == shortFlag || arg == longFlag {
			if i+1 < len(args) {
				next := args[i+1]
				if next != "" && !strings.HasPrefix(next, "-") {
					return next
				}
			}
			continue
		}
	}

	return ""
}

// HasSubCommandPath reports whether args contain a command path where `parent`
// is followed by one of the provided subcommands as the next non-flag token.
func HasSubCommandPath(args []string, parent string, subcommands ...string) bool {
	return HasSubCommandPathWithValueFlags(args, parent, defaultValueTakingFlags, subcommands...)
}

// HasSubCommandPathWithValueFlags is like HasSubCommandPath, but lets callers
// define which flags consume the next token as a value.
func HasSubCommandPathWithValueFlags(args []string, parent string, valueTakingFlags []string, subcommands ...string) bool {
	if len(subcommands) == 0 {
		return false
	}

	allowed := make(map[string]struct{}, len(subcommands))
	for _, sub := range subcommands {
		allowed[sub] = struct{}{}
	}

	consumesValue := make(map[string]struct{}, len(valueTakingFlags))
	for _, f := range valueTakingFlags {
		consumesValue[f] = struct{}{}
	}

	for i := range args {
		if args[i] == "--" {
			break
		}
		if args[i] != parent {
			continue
		}

		for j := i + 1; j < len(args); j++ {
			next := args[j]
			if next == "--" {
				break
			}
			if strings.HasPrefix(next, "-") {
				_, takesValue := consumesValue[next]
				if takesValue && !strings.Contains(next, "=") && j+1 < len(args) {
					value := args[j+1]
					if value != "" && !strings.HasPrefix(value, "-") {
						j++
					}
				}
				continue
			}
			_, ok := allowed[next]
			return ok
		}
	}

	return false
}

// PositionalArgAfterSubcommand returns the first positional argument after a
// specific subcommand, skipping known value-taking flags and their values.
func PositionalArgAfterSubcommand(args []string, subcommand string, valueTakingFlags []string) string {
	consumesValue := make(map[string]struct{}, len(valueTakingFlags))
	for _, f := range valueTakingFlags {
		consumesValue[f] = struct{}{}
	}

	seenSubcommand := false
	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == subcommand {
			seenSubcommand = true
			continue
		}
		if !seenSubcommand {
			continue
		}

		if arg == "--" {
			if i+1 < len(args) {
				return args[i+1]
			}
			return ""
		}

		for f := range consumesValue {
			if strings.HasPrefix(arg, f+"=") {
				goto nextArg
			}
		}

		if _, ok := consumesValue[arg]; ok {
			i++
			continue
		}

		if strings.HasPrefix(arg, "-") {
			continue
		}

		return arg

	nextArg:
	}

	return ""
}
