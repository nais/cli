package issues

import (
	"fmt"
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
			ret.Severity = gql.Severity(value)
		case "resourcename":
			ret.ResourceName = value
		case "resourcetype":
			ret.ResourceType = gql.ResourceType(value)
		case "issuetype":
			ret.IssueType = gql.IssueType(value)
		default:
			return gql.IssueFilter{}, fmt.Errorf("unknown filter key: %s", key)
		}

	}
	return ret, nil
}
