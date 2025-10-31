package issues

import (
	"fmt"
	"slices"
	"strings"

	"github.com/nais/cli/internal/naisapi/gql"
)

func ParseFilter(s string) (gql.IssueFilter, error) {
	if s == "" {
		return gql.IssueFilter{}, nil
	}
	ret := gql.IssueFilter{}
	parts := strings.SplitSeq(s, ",")
	for part := range parts {
		kv := strings.Split(part, "=")
		if len(kv) != 2 {
			return gql.IssueFilter{}, fmt.Errorf("incorrect filter: %s", part)
		}
		key, value := kv[0], kv[1]
		switch strings.ToLower(key) {
		case "environment":
			ret.Environments = []string{value}
		case "severity":
			s, err := parseEnumValue(value, gql.AllSeverity)
			if err != nil {
				return gql.IssueFilter{}, err
			}
			ret.Severity = s
		case "resourcename":
			ret.ResourceName = value
		case "resourcetype":
			rt, err := parseEnumValue(value, gql.AllResourceType)
			if err != nil {
				return gql.IssueFilter{}, err
			}
			ret.ResourceType = rt
		case "issuetype":
			it, err := parseEnumValue(value, gql.AllIssueType)
			if err != nil {
				return gql.IssueFilter{}, err
			}
			ret.IssueType = it
		default:
			return gql.IssueFilter{}, fmt.Errorf("unknown filter key: %s", key)
		}

	}
	return ret, nil
}

// parseEnumValue checks if value is valid for the given validValues slice and returns the value as the target type.
// If not valid, returns an error with the valid values listed.
func parseEnumValue[T ~string](value string, validValues []T) (T, error) {
	v := T(strings.ToUpper(value))
	if slices.Contains(validValues, v) {
		return v, nil
	}
	return "", fmt.Errorf("invalid filter value: %s, valid values are: %+v", value, validValues)
}
