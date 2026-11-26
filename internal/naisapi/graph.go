package naisapi

import (
	"context"

	"github.com/Khan/genqlient/graphql"
)

func GraphqlClient(ctx context.Context) (graphql.Client, error) {
	var user AuthenticatedUser
	if auth, ok := ctx.Value(invocationAuthKey{}).(*invocationAuth); ok {
		user = auth.user
	}

	if user == nil {
		var err error
		user, err = GetAuthenticatedUser(ctx)
		if err != nil {
			return nil, err
		}
	}

	return graphql.NewClient(user.APIURL(), user.HTTPClient(ctx)), nil
}
