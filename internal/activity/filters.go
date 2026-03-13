package activity

import (
	"fmt"
	"strings"

	"github.com/nais/cli/internal/naisapi/gql"
)

func ParseActivityTypes(in []string) ([]gql.ActivityLogActivityType, error) {
	return parseEnum[gql.ActivityLogActivityType](in, gql.AllActivityLogActivityType, "activity type")
}

func ParseResourceTypes(in []string) ([]gql.ActivityLogEntryResourceType, error) {
	return parseEnum[gql.ActivityLogEntryResourceType](in, gql.AllActivityLogEntryResourceType, "resource type")
}

func EnumStrings[T ~string](in []T) []string {
	ret := make([]string, len(in))
	for i, s := range in {
		ret[i] = string(s)
	}
	return ret
}

func parseEnum[T ~string](in []string, allowedValues []T, label string) ([]T, error) {
	ret := make([]T, 0, len(in))
	allowed := make(map[string]T, len(allowedValues))
	for _, v := range allowedValues {
		allowed[string(v)] = v
	}

	for _, t := range in {
		normalized := strings.ToUpper(strings.TrimSpace(t))
		v, ok := allowed[normalized]
		if !ok {
			return nil, fmt.Errorf("invalid %s %q", label, t)
		}
		ret = append(ret, v)
	}

	return ret, nil
}
