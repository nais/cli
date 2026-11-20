package issues

import (
	"fmt"
	"slices"
	"strings"

	"github.com/nais/cli/internal/issues/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
)

func ParseFilter(flags *flag.List) (gql.IssueFilter, error) {
	ret := gql.IssueFilter{}

	if flags.Environment != "" {
		ret.Environments = []string{string(flags.Environment)}
	}
	issueType, err := ParseIssueType(flags.IssueType)
	if err != nil {
		return gql.IssueFilter{}, err
	}
	ret.IssueType = issueType
	if flags.ResourceName != "" {
		ret.ResourceName = new(string(flags.ResourceName))
	}
	if flags.ResourceType != "" {
		rt, err := parseEnumValue(string(flags.ResourceType), gql.AllResourceType)
		if err != nil {
			return gql.IssueFilter{}, err
		}
		ret.ResourceType = new(rt)
	}
	if flags.Severity != "" {
		s, err := parseEnumValue(string(flags.Severity), gql.AllSeverity)
		if err != nil {
			return gql.IssueFilter{}, err
		}

		ret.Severity = new(s)
	}

	return ret, nil
}

func ParseIssueType(value flag.IssueType) (*gql.IssueType, error) {
	if value == "" {
		return nil, nil
	}
	issueType, err := parseEnumValue(string(value), gql.AllIssueType)
	if err != nil {
		return nil, err
	}
	return &issueType, nil
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
