package user

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type User struct {
	Name    string          `json:"name"`
	Email   string          `json:"email"`
	Source  *gql.UserSource `json:"source"`
	IsAdmin adminStatus     `heading:"Admin" json:"is_admin"`
}

type adminStatus bool

func (a adminStatus) String() string {
	if a {
		return "✓"
	}
	return ""
}

func GetUsers(ctx context.Context, filter *gql.UserFilter) ([]User, error) {
	_ = `# @genqlient
		# @genqlient(for: "UserFilter.isAdmin", omitempty: true)
		query ListUsers(
			$after: Cursor
			$filter: UserFilter
		) {
			users(first: 1000, after: $after, orderBy: {field: NAME, direction: ASC}, filter: $filter) {
				nodes { name email source isAdmin }
				pageInfo { hasNextPage endCursor }
			}
		}
	`
	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	users := make([]User, 0)
	var after *string
	for {
		resp, err := gql.ListUsers(ctx, client, after, filter)
		if err != nil {
			return nil, err
		}
		for _, u := range resp.Users.Nodes {
			users = append(users, User{Name: u.Name, Email: u.Email, Source: u.Source, IsAdmin: adminStatus(u.IsAdmin)})
		}
		if !resp.Users.PageInfo.HasNextPage {
			return users, nil
		}
		next := resp.Users.PageInfo.EndCursor
		if next == nil || *next == "" || after != nil && *next == *after {
			return nil, fmt.Errorf("user list has another page but no new cursor")
		}
		after = next
	}
}
