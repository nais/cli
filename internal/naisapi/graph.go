package naisapi

import (
	"context"
	"fmt"

	"github.com/Khan/genqlient/graphql"
)

func GraphqlClient(ctx context.Context) (graphql.Client, error) {
	user, err := GetAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	gqlClient := graphql.NewClient(fmt.Sprintf("https://%s/graphql", user.ConsoleHost), user.HTTPClient(ctx))
	return gqlClient, nil
}
