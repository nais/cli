package naisapi

import (
	"context"

	"github.com/Khan/genqlient/graphql"
)

func GraphqlClient(ctx context.Context) (graphql.Client, error) {
	user, err := GetAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	gqlClient := graphql.NewClient(user.APIURL(), user.HTTPClient(ctx))
	return gqlClient, nil
}
