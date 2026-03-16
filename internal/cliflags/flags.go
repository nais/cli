package cliflags

import "strings"

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
