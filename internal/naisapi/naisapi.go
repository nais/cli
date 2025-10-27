package naisapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/Khan/genqlient/graphql"
	logflag "github.com/nais/cli/internal/log/command/flag"
	"github.com/nais/cli/internal/naisapi/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
	"github.com/suessflorian/gqlfetch"
	"github.com/vektah/gqlparser/v2/gqlerror"
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

	schema, err := gqlfetch.BuildClientSchemaWithHeaders(ctx, user.APIURL(), headers, false)
	if err != nil {
		return "", err
	}

	// There's a bug that causes quadruple quotes, so we replace them with three:
	schema = strings.ReplaceAll(schema, `""""`, `"""`)

	return schema, nil
}

func StartProxy(ctx context.Context, out *naistrix.OutputWriter, flags *flag.Proxy) error {
	user, err := GetAuthenticatedUser(ctx)
	if err != nil {
		return err
	}

	// Setup reverse proxy to forward requests to the target server, but using a custom transport that authenticates the request
	target := &url.URL{
		Scheme: "https",
		Host:   user.ConsoleHost(),
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

	out.Println("Forwarding requests from", "http://"+flags.ListenAddr, "to", target.String())
	// Start the server
	http.Handle("/graphql", proxy)
	http.Handle("/", playground.Handler("Nais API playground", "/graphql"))
	if err := http.ListenAndServe(flags.ListenAddr, nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func GetUserTeams(ctx context.Context) ([]gql.UserTeamsMeUserTeamsTeamMemberConnectionNodesTeamMember, error) {
	_ = `# @genqlient
		query UserTeams {
			me {
				... on User {
					teams {
						nodes {
							role
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

	resp, err := gql.UserTeams(ctx, client)
	if err != nil {
		return nil, err
	}

	if u, ok := resp.Me.(*gql.UserTeamsMeUser); ok {
		return u.Teams.Nodes, nil
	}

	return nil, nil
}

func GetStatus(ctx context.Context, _ *flag.Status) ([]gql.TeamStatusMeUserTeamsTeamMemberConnectionNodesTeamMember, error) {
	_ = `# @genqlient
		query TeamStatus {
			me {
				... on User {
					teams {
						nodes {
							team {
								slug
								workloads(first: 500) {
									nodes {
										__typename
										name
										teamEnvironment { environment { name } }
										issues(first: 100, filter: { severity: CRITICAL }) {
											nodes { __typename }
											pageInfo { totalCount }
										}
									}
									pageInfo { totalCount }
								}
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

	resp, err := gql.TeamStatus(ctx, client)
	if err != nil {
		return nil, err
	}

	if u, ok := resp.Me.(*gql.TeamStatusMeUser); ok {
		return u.Teams.Nodes, nil
	}

	return nil, nil
}

func GetAllTeamSlugs(ctx context.Context) ([]string, error) {
	ret, err := GetAllTeams(ctx)
	if err != nil {
		return nil, err
	}

	slugs := make([]string, len(ret.Teams.Nodes))
	for i, t := range ret.Teams.Nodes {
		slugs[i] = t.Slug
	}
	return slugs, nil
}

func GetAllTeams(ctx context.Context) (*gql.TeamsResponse, error) {
	_ = `# @genqlient
		query Teams {
			teams(first: 1000) {
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

// IsConsoleAdmin checks if the authenticated user is a Console admin or not.
func IsConsoleAdmin(ctx context.Context) bool {
	_ = `# @genqlient
		query IsAdmin {
			me {
				... on User { isAdmin }
			}
		}
	`

	client, err := GraphqlClient(ctx)
	if err != nil {
		return false
	}

	resp, err := gql.IsAdmin(ctx, client)
	if err != nil {
		return false
	}

	if u, ok := resp.Me.(*gql.IsAdminMeUser); ok {
		return u.IsAdmin
	}

	return false
}

func GetTeamMembers(ctx context.Context, teamSlug string) ([]gql.TeamMembersTeamMembersTeamMemberConnectionNodesTeamMember, error) {
	_ = `# @genqlient
		query TeamMembers($slug: Slug!) {
			team(slug: $slug) {
				members(first: 1000) {
					nodes {
						role
						user {
							name
							email
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

	resp, err := gql.TeamMembers(ctx, client, teamSlug)
	if err != nil {
		return nil, err
	}

	return resp.Team.Members.Nodes, nil
}

func GetTeamWorkloads(ctx context.Context, teamSlug string) ([]gql.GetTeamWorkloadsTeamWorkloadsWorkloadConnectionNodesWorkload, error) {
	_ = `# @genqlient
		query GetTeamWorkloads($slug: Slug!) {
			team(slug: $slug) {
				workloads(first: 1000) {
					nodes {
						__typename
						name
						... on Application {
						  applicationState: state
						}
						... on Job {
						  jobState: state
						}
						totalIssues: issues {
						  pageInfo {
							totalCount
						  }
						}
						image { vulnerabilitySummary { total } }
						teamEnvironment { environment { name } }
					}
				}
			}
		}
	`

	client, err := GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := gql.GetTeamWorkloads(ctx, client, teamSlug)
	if err != nil {
		return nil, err
	}

	return resp.Team.Workloads.Nodes, nil
}

func GetUserEmails(ctx context.Context) ([]string, error) {
	ret, err := GetUsers(ctx)
	if err != nil {
		return nil, err
	}

	emails := make([]string, len(ret.Users.Nodes))
	for i, u := range ret.Users.Nodes {
		emails[i] = u.Email
	}
	return emails, nil
}

func GetUsers(ctx context.Context) (*gql.UsersResponse, error) {
	_ = `# @genqlient
		query Users {
			users(first: 5000) {
				nodes {
					name
					email
				}
			}
		}
	`

	client, err := GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	return gql.Users(ctx, client)
}

func AddTeamMember(ctx context.Context, teamSlug, email string, role gql.TeamMemberRole) error {
	_ = `# @genqlient
		mutation AddTeamMember(
			$slug: Slug!
			$email: String!
			$role: TeamMemberRole!
		) {
			addTeamMember(input: {
				teamSlug: $slug
				userEmail: $email
				role: $role
			}) {
				member { role }
			}
		}
	`

	client, err := GraphqlClient(ctx)
	if err != nil {
		return err
	}

	_, err = gql.AddTeamMember(ctx, client, teamSlug, email, role)
	return err
}

func RemoveTeamMember(ctx context.Context, teamSlug, email string) error {
	_ = `# @genqlient
		mutation RemoveTeamMember(
			$slug: Slug!
			$email: String!
		) {
			removeTeamMember(input: {
				teamSlug: $slug
				userEmail: $email
			}) {
				team { slug }
			}
		}
	`

	client, err := GraphqlClient(ctx)
	if err != nil {
		return err
	}

	_, err = gql.RemoveTeamMember(ctx, client, teamSlug, email)
	return err
}

func TailLog(ctx context.Context, out *naistrix.OutputWriter, flag *logflag.LogFlags, lokiQuery string) error {
	gqlQuery := `# @genqlient
		subscription TailLog($environment: String!, $query: String!, $limit: Int, $start: Time) {
			log(
				filter: {
					environmentName: $environment
					query: $query
					initialBatch: {
						limit: $limit
						start: $start
					}
				}
			) {
				time
				message
				labels {
					key
					value
				}
			}
		}
	`

	req := graphql.Request{
		OpName: "TailLog",
		Query:  gqlQuery,
		Variables: struct {
			Environment string `json:"environment"`
			Query       string `json:"query"`
			Limit       int    `json:"limit"`
			Start       string `json:"start"`
		}{
			Environment: flag.Environment,
			Query:       lokiQuery,
			Limit:       flag.Limit,
			Start:       time.Now().Add(-flag.Since).Format(time.RFC3339),
		},
	}

	onData := func(entry gql.TailLogResponse) {
		if flag.WithTimestamps {
			out.Printf("%s: ", entry.Log.Time.Format(time.RFC3339))
		}
		out.Println(entry.Log.Message)
		if flag.WithLabels {
			var labels []string
			for _, label := range entry.Log.Labels {
				labels = append(labels, label.Key+"="+label.Value)
			}
			out.Printf("Labels: [%s]\n", strings.Join(labels, ", "))
		}
	}

	onError := func(err gqlerror.Error) {
		out.Printf("Error: %v", err)
	}

	return SSEQuery(ctx, req, onData, onError)
}
