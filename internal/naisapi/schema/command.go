package schema

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/command/flag"
	"github.com/suessflorian/gqlfetch"
)

func Pull(ctx context.Context, _ *flag.Schema) (string, error) {
	user, err := naisapi.GetAuthenticatedUser(ctx)
	if err != nil {
		return "", err
	}

	headers := http.Header{}
	err = user.SetAuthorizationHeader(headers)
	if err != nil {
		return "", err
	}

	schema, err := gqlfetch.BuildClientSchemaWithHeaders(ctx, fmt.Sprintf("https://%s/graphql", user.ConsoleHost), headers, false)
	if err != nil {
		return "", err
	}

	// There's a bug that causes quadruple quotes, so we replace them with three:
	schema = strings.ReplaceAll(schema, `""""`, `"""`)

	return schema, nil
}
