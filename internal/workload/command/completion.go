package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/workload"
	"github.com/nais/naistrix"
)

func CompleteWorkloadName(ctx context.Context, args *naistrix.Arguments, prefix string, kind gql.SearchType, team, environment string) ([]string, string) {
	if args.Len() != 0 {
		return nil, ""
	}
	if prefix == "" {
		return nil, "Type part of a workload name to search."
	}
	names, err := workload.SearchNames(ctx, kind, team, environment, prefix)
	if err != nil {
		return nil, fmt.Sprintf("Unable to search workload names: %v", err)
	}
	return names, ""
}
