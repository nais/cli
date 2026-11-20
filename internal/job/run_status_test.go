package job

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
