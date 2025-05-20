package schema

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/nais/cli/internal/naisapi"
	"github.com/suessflorian/gqlfetch"
)

type Flags struct {
	*naisapi.Flags
}

func Pull(ctx context.Context, _ *Flags) (string, error) {
	secret, err := naisapi.GetUserSecret(ctx)
	if err != nil {
		return "", err
	}

	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+secret.AccessToken)

	schema, err := gqlfetch.BuildClientSchemaWithHeaders(ctx, fmt.Sprintf("https://%s/graphql", secret.ConsoleHost), headers, false)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// There's a bug that causes quadruple quotes, so we replace them with three:
	schema = strings.ReplaceAll(schema, `""""`, `"""`)

	return schema, nil
}
