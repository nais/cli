package job

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/nais/cli/internal/naisapi/gql"
)

func TestGetRunStatusSelectsRun(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			OperationName string `json:"operationName"`
			Variables     struct {
				Job, Env, After, ID string
				First               int
			}
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		v := request.Variables
		switch request.OperationName {
		case "FindJobRuns":
			if v.Job != "batch" || v.Env != "dev" {
				t.Errorf("looking up runs for the wrong job or environment: %+v", v)
			}
			if v.First == 1 && v.After != "" {
				t.Errorf("unexpected cursor for latest run: %q", v.After)
			}
			if v.After == "" {
				_, _ = io.WriteString(w, `{"data":{"team":{"environment":{"job":{"runs":{"nodes":[{"id":"JR_new","name":"new"}],"pageInfo":{"hasNextPage":true,"endCursor":"next-page"}}}}}}}`)
				return
			}
			if v.After != "next-page" {
				t.Errorf("unexpected run cursor: %q", v.After)
			}
			_, _ = io.WriteString(w, `{"data":{"team":{"environment":{"job":{"runs":{"nodes":[{"id":"JR_old","name":"old"}],"pageInfo":{"hasNextPage":false}}}}}}}`)
		case "GetJobRunDetails":
			if v.ID == "JR_new" {
				_, _ = io.WriteString(w, `{"data":{"node":{"__typename":"JobRun","name":"new","status":{"state":"RUNNING"},"instances":{"nodes":[],"pageInfo":{"hasNextPage":false}}}}}`)
				return
			}
			if v.ID != "JR_old" {
				t.Errorf("requested run ID %q, want JR_old", v.ID)
			}
			_, _ = io.WriteString(w, `{"data":{"node":{"__typename":"JobRun","name":"old","status":{"state":"FAILED","message":"Run failed."},"instances":{"nodes":[{"name":"instance-1","image":{"name":"image","tag":"old"},"status":{"state":"FAILING","message":"Instance failed.","lastExitCode":1}}],"pageInfo":{"hasNextPage":false}}}}}`)
		default:
			t.Errorf("unexpected operation %q", request.OperationName)
			http.Error(w, "unexpected operation", http.StatusBadRequest)
		}
	}))
	defer server.Close()
	t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))

	latest, err := GetRunStatus(t.Context(), "team", "dev", "batch", "")
	if err != nil {
		t.Fatal(err)
	}
	if latest.Run != "new" || latest.State != gql.JobRunStateRunning {
		t.Fatalf("got %+v, want the latest run", latest)
	}

	got, err := GetRunStatus(t.Context(), "team", "dev", "batch", "old")
	if err != nil {
		t.Fatal(err)
	}
	if got.State != gql.JobRunStateFailed || len(got.Instances) != 1 || got.Instances[0].State != gql.JobRunInstanceStateFailing {
		t.Fatalf("got %+v, want a failed run with a failing instance", got)
	}
}

func TestGetRunStatusPaginatesInstances(t *testing.T) {
	const cursor = "instances-2"
	type pageInfo struct {
		HasNextPage bool   `json:"hasNextPage"`
		EndCursor   string `json:"endCursor,omitempty"`
	}
	next := pageInfo{HasNextPage: true, EndCursor: cursor}
	for _, tt := range []struct {
		name    string
		pages   []pageInfo
		wantErr bool
	}{
		{name: "accumulates instances", pages: []pageInfo{next, {}}},
		{name: "missing cursor", pages: []pageInfo{{HasNextPage: true}}, wantErr: true},
		{name: "repeated cursor", pages: []pageInfo{next, next}, wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var detailCalls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					OperationName string
					Variables     struct {
						ID    string
						After *string
					}
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Error(err)
					http.Error(w, "invalid request", http.StatusBadRequest)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				switch request.OperationName {
				case "FindJobRuns":
					_, _ = io.WriteString(w, `{"data":{"team":{"environment":{"job":{"runs":{"nodes":[{"id":"JR_new","name":"new"}],"pageInfo":{"hasNextPage":false}}}}}}}`)
				case "GetJobRunDetails":
					page := int(detailCalls.Add(1)) - 1
					if page >= len(tt.pages) {
						t.Error("unexpected instance page request")
						http.Error(w, "unexpected instance page", http.StatusBadRequest)
						return
					}
					if request.Variables.ID != "JR_new" {
						t.Errorf("run ID = %q, want JR_new", request.Variables.ID)
					}
					after := request.Variables.After
					if page == 0 && after != nil {
						t.Errorf("first page cursor = %q, want nil", *after)
					}
					if page == 1 && (after == nil || *after != cursor) {
						t.Errorf("second page request did not use cursor %q", cursor)
					}
					info, _ := json.Marshal(tt.pages[page])
					_, _ = fmt.Fprintf(w, `{"data":{"node":{"__typename":"JobRun","name":"new","status":{"state":"RUNNING"},"instances":{
						"nodes":[{"name":"instance-%d","image":{"name":"image","tag":"latest"},"status":{"state":"RUNNING"}}],
						"pageInfo":%s
					}}}}`, page+1, info)
				default:
					t.Errorf("unexpected operation %q", request.OperationName)
					http.Error(w, "unexpected operation", http.StatusBadRequest)
				}
			}))
			defer server.Close()
			t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))

			got, err := GetRunStatus(t.Context(), "team", "dev", "batch", "")
			if calls := int(detailCalls.Load()); calls != len(tt.pages) {
				t.Errorf("requested %d instance pages, want %d", calls, len(tt.pages))
			}
			if tt.wantErr {
				const want = `paging instances for run "new": API reported more results without a new cursor`
				if err == nil || err.Error() != want || got != nil {
					t.Fatalf("got status %+v, error %v; want no status and error %q", got, err, want)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var names []string
			for _, instance := range got.Instances {
				names = append(names, instance.Name)
			}
			if want := []string{"instance-1", "instance-2"}; !slices.Equal(names, want) {
				t.Errorf("instances = %v, want %v", names, want)
			}
		})
	}
}
