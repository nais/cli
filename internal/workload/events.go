package workload

import (
	"context"
	"fmt"
	"time"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

var _ = `# @genqlient
	fragment WorkloadEventsPage on WorkloadEventConnection {
		pageInfo { hasNextPage endCursor }
		nodes {
			id
			message
			triggeredAt
			... on WorkloadWarningEvent { kind instance count firstSeenAt }
			... on WorkloadKillingEvent { container reason }
			... on WorkloadSuccessfulRescaleEvent { newSize direction }
		}
	}

	query GetApplicationEvents($team: Slug!, $env: String!, $name: String!, $limit: Int!, $after: Cursor, $filter: WorkloadEventFilter) {
		team(slug: $team) {
			environment(name: $env) {
				application(name: $name) {
					# @genqlient(flatten: true)
					events(first: $limit, after: $after, filter: $filter) {
						...WorkloadEventsPage
					}
				}
			}
		}
	}

	query GetJobEvents($team: Slug!, $env: String!, $name: String!, $limit: Int!, $after: Cursor, $filter: WorkloadEventFilter) {
		team(slug: $team) {
			environment(name: $env) {
				job(name: $name) {
					# @genqlient(flatten: true)
					events(first: $limit, after: $after, filter: $filter) {
						...WorkloadEventsPage
					}
				}
			}
		}
	}
`

// EventsPage keeps the query bounds with the cursor, so callers can page within the same time range.
type EventsPage struct {
	Events    []Event    `json:"events"`
	From      *time.Time `json:"from,omitempty"`
	To        time.Time  `json:"to"`
	NextAfter *string    `json:"nextAfter,omitempty"`
}

type Event struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"`
	Kind        string     `json:"kind,omitempty"`
	Message     string     `json:"message"`
	TriggeredAt time.Time  `json:"triggeredAt"`
	Instance    *string    `json:"instance,omitempty"`
	Count       int        `json:"count,omitempty"`
	FirstSeenAt *time.Time `json:"firstSeenAt,omitempty"`
	Container   string     `json:"container,omitempty"`
	Reason      string     `json:"reason,omitempty"`
	NewSize     *int       `json:"newSize,omitempty"`
	Direction   string     `json:"direction,omitempty"`
}

type EventQuery struct {
	Limit int
	After *string
	From  *time.Time
	To    time.Time
}

type EventSource string

const (
	ApplicationEvents EventSource = "application"
	JobEvents         EventSource = "job"
)

func GetEvents(ctx context.Context, source EventSource, team, environment, name string, query EventQuery) (*EventsPage, error) {
	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}
	filter := &gql.WorkloadEventFilter{From: query.From, To: &query.To}
	var page gql.WorkloadEventsPage
	switch source {
	case ApplicationEvents:
		resp, err := gql.GetApplicationEvents(ctx, client, team, environment, name, query.Limit, query.After, filter)
		if err != nil {
			return nil, err
		}
		page = resp.Team.Environment.Application.Events
	case JobEvents:
		resp, err := gql.GetJobEvents(ctx, client, team, environment, name, query.Limit, query.After, filter)
		if err != nil {
			return nil, err
		}
		page = resp.Team.Environment.Job.Events
	default:
		return nil, fmt.Errorf("unknown workload event source %q", source)
	}
	seen := make(map[string]bool)
	if query.After != nil {
		seen[*query.After] = true
	}
	next, err := NextCursor(page.PageInfo.HasNextPage, page.PageInfo.EndCursor, seen)
	if err != nil {
		return nil, err
	}
	ret := &EventsPage{
		Events: make([]Event, 0, len(page.Nodes)), From: query.From, To: query.To, NextAfter: next,
	}
	for _, node := range page.Nodes {
		event := Event{ID: node.GetId(), Message: node.GetMessage(), TriggeredAt: node.GetTriggeredAt()}
		switch n := node.(type) {
		case *gql.WorkloadEventsPageNodesWorkloadWarningEvent:
			event.Type, event.Kind, event.Instance = "warning", string(n.Kind), n.Instance
			event.Count, event.FirstSeenAt = n.Count, &n.FirstSeenAt
		case *gql.WorkloadEventsPageNodesWorkloadKillingEvent:
			event.Type, event.Container, event.Reason = "restart", n.Container, n.Reason
		case *gql.WorkloadEventsPageNodesWorkloadSuccessfulRescaleEvent:
			event.Type, event.NewSize, event.Direction = "rescale", &n.NewSize, string(n.Direction)
		}
		ret.Events = append(ret.Events, event)
	}
	return ret, nil
}
