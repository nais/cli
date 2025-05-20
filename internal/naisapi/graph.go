package naisapi

import (
	"context"
	"fmt"

	"github.com/Khan/genqlient/graphql"
)

func graphqlClient(ctx context.Context) (graphql.Client, error) {
	client, consoleHost, err := AuthenticatedHTTPClient(ctx)
	if err != nil {
		return nil, err
	}

	gqlClient := graphql.NewClient(fmt.Sprintf("https://%s/graphql", consoleHost), client)
	return gqlClient, nil
}
