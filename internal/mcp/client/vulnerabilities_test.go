package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Khan/genqlient/graphql"
)

type vulnerabilityTransport func(*http.Request) (*http.Response, error)

func (f vulnerabilityTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func vulnerabilityHTTPClient(t *testing.T, response map[string]any, inspect func(*http.Request)) graphql.Client {
	t.Helper()
	body, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	return graphql.NewClient("https://nais.example.test/graphql", &http.Client{Transport: vulnerabilityTransport(func(r *http.Request) (*http.Response, error) {
		if inspect != nil {
			inspect(r)
		}
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(string(body)))}, nil
	})})
}

func vulnerabilityResponse(kind string) map[string]any {
	node := map[string]any{
		"id": "workload-1", "name": "app",
		"teamEnvironment":  map[string]any{"environment": map[string]any{"name": "prod"}},
		"ingresses":        []any{map[string]any{"url": "https://app.example.test", "type": "EXTERNAL"}},
		"authIntegrations": []any{map[string]any{"__typename": "EntraIDAuthIntegration"}},
		"networkPolicy":    map[string]any{"inbound": map[string]any{"rules": []any{map[string]any{"targetWorkloadName": "caller", "targetTeamSlug": "other-team", "mutual": true}}}},
		"image": map[string]any{
			"id": "image-1", "name": "example/app", "tag": "", "digest": "sha256:example",
			"sbom":                 map[string]any{"status": "READY"},
			"vulnerabilitySummary": map[string]any{"total": 0, "lastUpdated": "2026-01-01T00:00:00Z"},
			"vulnerabilities": map[string]any{
				"pageInfo": map[string]any{"hasNextPage": false, "endCursor": "cursor-1"},
				"nodes": []any{map[string]any{
					"id": "finding-1", "identifier": "CVE-2025-0001", "package": "pkg:npm/example@1.0", "severity": "HIGH", "description": "Untrusted input can crash a parser.",
					"riskAssessment":           map[string]any{"cvssScore": 7.5, "epssScore": 0.15, "epssPercentile": 95.0, "hasKevEntry": true, "knownRansomwareUse": false},
					"remediation":              map[string]any{"fixVersion": nil, "latestVersion": "2.0"},
					"suppression":              map[string]any{"state": "IN_TRIAGE", "reason": "must not be returned"},
					"vulnerabilityDetailsLink": "https://example.test/advisory",
				}},
			},
		},
	}
	return map[string]any{"data": map[string]any{"team": map[string]any{
		"slug":        "team",
		"environment": map[string]any{kind: node},
	}}}
}

func vulnerabilityResponseNode(response map[string]any, kind string) map[string]any {
	return response["data"].(map[string]any)["team"].(map[string]any)["environment"].(map[string]any)[kind].(map[string]any)
}

func TestFetchWorkloadVulnerabilitiesOverlappingNames(t *testing.T) {
	for _, kind := range []string{"application", "job"} {
		t.Run(kind, func(t *testing.T) {
			response := vulnerabilityResponse(kind)
			node := vulnerabilityResponseNode(response, kind)
			node["name"] = "api"
			team := response["data"].(map[string]any)["team"].(map[string]any)
			team[kind+"s"] = map[string]any{
				"pageInfo": map[string]any{"hasNextPage": false},
				"nodes":    []any{node, map[string]any{"id": "other-workload", "name": "api-gateway"}},
			}
			gqlClient := vulnerabilityHTTPClient(t, response, func(r *http.Request) {
				var request struct {
					Query     string         `json:"query"`
					Variables map[string]any `json:"variables"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(request.Query, "environment(name: $env)") ||
					!strings.Contains(request.Query, kind+"(name: $name)") ||
					strings.Contains(request.Query, kind+"s(") ||
					request.Variables["name"] != "api" {
					t.Fatalf("expected scoped exact-name lookup, got %s", request.Query)
				}
			})
			page, err := fetchWorkloadVulnerabilities(context.Background(), gqlClient, WorkloadVulnerabilityRequest{
				Team: "team", Workload: "api", Environment: "prod", Type: kind, First: 10,
			})
			if err != nil {
				t.Fatal(err)
			}
			if page.Workload.ID != "workload-1" || page.Workload.Name != "api" || len(page.Findings) != 1 {
				t.Fatalf("wrong exact workload selected: %+v", page)
			}
		})
	}
}

func TestFetchWorkloadVulnerabilities(t *testing.T) {
	for _, kind := range []string{"application", "job"} {
		t.Run(kind, func(t *testing.T) {
			response := vulnerabilityResponse(kind)
			gqlClient := vulnerabilityHTTPClient(t, response, func(r *http.Request) {
				var request struct {
					Query     string         `json:"query"`
					Variables map[string]any `json:"variables"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Error(err)
					return
				}
				if request.Variables["team"] != "team" || request.Variables["name"] != "app" || request.Variables["env"] != "prod" || request.Variables["first"] != float64(10) || request.Variables["after"] != "cursor-0" {
					t.Errorf("scope/pagination variables lost: %+v", request.Variables)
				}
				for _, field := range []string{"priority", "manifest", "secrets", "reason", "hasSBOM"} {
					if strings.Contains(request.Query, field) {
						t.Errorf("query contains excluded field %s", field)
					}
				}
			})
			cursor := "cursor-0"
			page, err := fetchWorkloadVulnerabilities(context.Background(), gqlClient, WorkloadVulnerabilityRequest{
				Team: "team", Workload: "app", Environment: "prod", Type: kind, First: 10, After: &cursor,
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(page.Findings) != 1 || page.Findings[0].Remediation.FixVersion != nil || *page.Findings[0].RiskAssessment.EPSSPercentile != 95 || *page.Findings[0].RiskAssessment.EPSSScore != 0.15 {
				t.Fatalf("risk/null/version data lost: %+v", page)
			}
			if page.Findings[0].SuppressionState == nil || *page.Findings[0].SuppressionState != "IN_TRIAGE" || page.Snapshot.ImageTag != "" || page.Snapshot.ImageDigest == nil {
				t.Fatal("suppression or digest-only image was lost")
			}
			if page.Snapshot.SummaryUpdatedAt == nil || page.Snapshot.SummaryUpdatedAt.Format(time.RFC3339) != "2026-01-01T00:00:00Z" {
				t.Fatal("summary update timestamp was lost")
			}
			if len(page.Context.InboundRules) != 1 || page.Context.InboundRules[0].Team != "other-team" || len(page.Context.AuthIntegrations) != 1 {
				t.Fatalf("context lost: %+v", page.Context)
			}
			if kind == "application" && len(page.Context.Ingresses) != 1 || kind == "job" && len(page.Context.Ingresses) != 0 {
				t.Fatalf("wrong ingress context for %s", kind)
			}
			raw, err := json.Marshal(page)
			if err != nil || strings.Contains(string(raw), "must not be returned") {
				t.Fatal("suppression free text escaped DTO")
			}
			if strings.Contains(string(raw), `"scan_updated_at"`) {
				t.Fatal("summary update exposed as a scan timestamp")
			}
		})
	}
}

func TestFetchWorkloadVulnerabilitiesRejectsScopeAndErrors(t *testing.T) {
	for _, test := range []struct {
		name string
		edit func(map[string]any)
		want string
	}{
		{"wrong team", func(r map[string]any) { r["data"].(map[string]any)["team"].(map[string]any)["slug"] = "other" }, "scope_mismatch"},
		{"no workload", func(r map[string]any) {
			delete(r["data"].(map[string]any)["team"].(map[string]any)["environment"].(map[string]any), "application")
		}, "scope_mismatch"},
		{"wrong workload", func(r map[string]any) {
			vulnerabilityResponseNode(r, "application")["name"] = "app-gateway"
		}, "scope_mismatch"},
		{"wrong environment", func(r map[string]any) {
			node := vulnerabilityResponseNode(r, "application")
			node["teamEnvironment"] = map[string]any{"environment": map[string]any{"name": "dev"}}
		}, "scope_mismatch"},
		{"GraphQL errors with partial data", func(r map[string]any) { r["errors"] = []any{map[string]any{"message": "access denied"}} }, "access denied"},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := vulnerabilityResponse("application")
			test.edit(response)
			_, err := fetchWorkloadVulnerabilities(context.Background(), vulnerabilityHTTPClient(t, response, nil), WorkloadVulnerabilityRequest{Team: "team", Workload: "app", Environment: "prod", Type: "application", First: 10})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("err=%v, want %s", err, test.want)
			}
		})
	}
}

func TestFetchWorkloadVulnerabilitiesWithoutSummary(t *testing.T) {
	response := vulnerabilityResponse("job")
	node := vulnerabilityResponseNode(response, "job")
	node["image"].(map[string]any)["vulnerabilitySummary"] = nil
	page, err := fetchWorkloadVulnerabilities(context.Background(), vulnerabilityHTTPClient(t, response, nil), WorkloadVulnerabilityRequest{Team: "team", Workload: "app", Environment: "prod", Type: "job", First: 10})
	if err != nil {
		t.Fatal(err)
	}
	if page.Snapshot.Total != nil || page.Snapshot.SummaryUpdatedAt != nil || len(page.Findings) != 1 {
		t.Fatalf("missing summary interpreted as clean: %+v", page)
	}
}
