package workload

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/nais/cli/internal/naisapi/gql"
)

func TestGetEventsDecodesWorkloadEventTypes(t *testing.T) {
	from := time.Date(2026, 11, 1, 10, 0, 0, 123456789, time.UTC)
	to := from.Add(3 * time.Hour)
	firstSeen := from.Add(30 * time.Minute)
	const events = `{"pageInfo":{"hasNextPage":true,"endCursor":"next-cursor"},"nodes":[
		{"__typename":"WorkloadWarningEvent","id":"warning-id","message":"Could not start instance","triggeredAt":"2026-11-01T12:00:00.123456789Z","kind":"START_FAILED","instance":"pod-1","count":3,"firstSeenAt":"2026-11-01T10:30:00.123456789Z"},
		{"__typename":"WorkloadKillingEvent","id":"restart-id","message":"Restarted container","triggeredAt":"2026-11-01T11:00:00.123456789Z","container":"worker","reason":"Liveness probe failed"},
		{"__typename":"WorkloadSuccessfulRescaleEvent","id":"rescale-id","message":"Scaled to zero","triggeredAt":"2026-11-01T10:00:00.123456789Z","newSize":0,"direction":"DOWN"}
	]}`
	for _, tt := range []struct {
		source    EventSource
		operation string
	}{
		{ApplicationEvents, "GetApplicationEvents"},
		{JobEvents, "GetJobEvents"},
	} {
		t.Run(string(tt.source), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					OperationName string
					Variables     struct {
						Team, Env, Name, After string
						Limit                  int
						Filter                 gql.WorkloadEventFilter
					}
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Errorf("decoding request: %v", err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				v := request.Variables
				if request.OperationName != tt.operation || v.Team != "my-team" || v.Env != "dev" || v.Name != "worker" || v.After != "previous-cursor" || v.Limit != 3 {
					t.Errorf("unexpected events request: %+v", request)
				}
				if diff := cmp.Diff(gql.WorkloadEventFilter{From: &from, To: &to}, v.Filter); diff != "" {
					t.Errorf("events filter (-want +got):\n%s", diff)
				}
				w.Header().Set("Content-Type", "application/json")
				if _, err := io.WriteString(w, `{"data":{"team":{"environment":{"`+string(tt.source)+`":{"events":`+events+`}}}}}`); err != nil {
					t.Errorf("writing response: %v", err)
				}
			}))
			defer server.Close()
			t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))

			got, err := GetEvents(t.Context(), tt.source, "my-team", "dev", "worker", EventQuery{
				Limit: 3, After: new("previous-cursor"), From: &from, To: to,
			})
			if err != nil {
				t.Fatal(err)
			}
			want := &EventsPage{
				From: &from, To: to, NextAfter: new("next-cursor"),
				Events: []Event{
					{ID: "warning-id", Type: "warning", Kind: "START_FAILED", Message: "Could not start instance", TriggeredAt: from.Add(2 * time.Hour), Instance: new("pod-1"), Count: 3, FirstSeenAt: &firstSeen},
					{ID: "restart-id", Type: "restart", Message: "Restarted container", TriggeredAt: from.Add(time.Hour), Container: "worker", Reason: "Liveness probe failed"},
					{ID: "rescale-id", Type: "rescale", Message: "Scaled to zero", TriggeredAt: from, NewSize: new(0), Direction: "DOWN"},
				},
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("events page (-want +got):\n%s", diff)
			}
		})
	}
}
