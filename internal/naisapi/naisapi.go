package naisapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/nais/cli/internal/naisapi/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/output"
	"github.com/suessflorian/gqlfetch"
)

func PullSchema(ctx context.Context, _ *flag.Schema) (string, error) {
	user, err := GetAuthenticatedUser(ctx)
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

func StartProxy(ctx context.Context, out output.Output, flags *flag.Proxy) error {
	user, err := GetAuthenticatedUser(ctx)
	if err != nil {
		return err
	}

	// Setup reverse proxy to forward requests to the target server, but using a custom transport that authenticates the request
	target := &url.URL{
		Scheme: "https",
		Host:   user.ConsoleHost,
	}
	proxy := &httputil.ReverseProxy{
		Rewrite: func(req *httputil.ProxyRequest) {
			req.SetURL(target)
			req.Out.Header.Set("user-agent", req.In.Header.Get("user-agent")+" (nais-cli)")
		},
		Transport: user.RoundTripper(&http.Transport{
			Proxy: http.ProxyFromEnvironment,
		}),
	}

	out.Println("Forwarding requests from", flags.ListenAddr, "to", target.String())
	// Start the server
	http.Handle("/", proxy)
	if err := http.ListenAndServe(flags.ListenAddr, nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func GetUserTeams(ctx context.Context, _ *flag.Teams) (*gql.UserTeamsResponse, error) {
	_ = `# @genqlient
		query UserTeams {
			me {
				... on User {
					teams {
						nodes {
							team {
								slug
								purpose
							}
						}
					}
				}
			}
		}
	`

	client, err := GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	return gql.UserTeams(ctx, client)
}

func GetAllTeams(ctx context.Context) (*gql.TeamsResponse, error) {
	_ = `# @genqlient
		query Teams {
	    teams(first:1000) {
				nodes {
					slug
					purpose
				}
			}
		}
	`

	client, err := GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	return gql.Teams(ctx, client)
}
