package tools

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"testing"

	"github.com/Khan/genqlient/graphql"
)

type reportTestVariables struct {
	Team         string   `json:"team"`
	Name         string   `json:"name"`
	Environment  string   `json:"environment"`
	Environments []string `json:"environments"`
	After        *string  `json:"after"`
	Applications bool     `json:"applications"`
	Jobs         bool     `json:"jobs"`
}

type reportTestClient func(context.Context, string, reportTestVariables) (any, error)

func (c reportTestClient) MakeRequest(ctx context.Context, req *graphql.Request, resp *graphql.Response) error {
	data, err := json.Marshal(req.Variables)
	if err != nil {
		return err
	}
	var variables reportTestVariables
	if err := json.Unmarshal(data, &variables); err != nil {
		return err
	}
	result, err := c(ctx, req.OpName, variables)
	if err != nil {
		return err
	}
	data, err = json.Marshal(result)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, resp.Data)
}

func reportTestNode(name, environment, imageName, tag string, digest *string) map[string]any {
	return map[string]any{
		"name": name, "teamEnvironment": map[string]any{"environment": map[string]any{"name": environment}},
		"image": map[string]any{
			"id": "image-id", "name": imageName, "tag": tag, "digest": digest,
			"sbom": map[string]any{"status": "READY"},
			"vulnerabilitySummary": map[string]any{
				"total": 3, "lastUpdated": "2026-11-16T09:00:00Z",
			},
		},
		"deployments": map[string]any{"nodes": []any{
			map[string]any{"repository": "nais/oracleverk", "commitSha": "deployed-commit", "createdAt": "2026-09-14T20:03:00Z"},
		}},
	}
}

func reportTestConnection(nodes []any, next bool, cursor *string) map[string]any {
	return map[string]any{"nodes": nodes, "pageInfo": map[string]any{"hasNextPage": next, "endCursor": cursor}}
}

func reportTestFindingPage(kind, imageName, tag string, digest *string, nodes []any, next bool, cursor *string) map[string]any {
	return map[string]any{"team": map[string]any{"environment": map[string]any{kind: map[string]any{
		"image": map[string]any{
			"name": imageName, "tag": tag, "digest": digest,
			"vulnerabilities": reportTestConnection(nodes, next, cursor),
		},
	}}}}
}

func TestWorkloadVulnerabilityReport(t *testing.T) {
	for _, kind := range []string{"application", "job"} {
		t.Run(kind, func(t *testing.T) {
			digest := "sha256:shared"
			discoveryPages, findingPages := 0, 0
			client := reportTestClient(func(_ context.Context, operation string, vars reportTestVariables) (any, error) {
				if vars.Team != "nais" || vars.Name != "oracleverk" {
					t.Fatalf("unexpected target: %+v", vars)
				}
				if operation == "MCPVulnerabilityWorkloads" {
					connection := "applications"
					if vars.Jobs {
						connection = "jobs"
					}
					if vars.Applications == vars.Jobs {
						t.Fatalf("expected exactly one workload kind: %+v", vars)
					}
					if connection != kind+"s" {
						return map[string]any{"team": map[string]any{connection: reportTestConnection([]any{}, false, nil)}}, nil
					}
					discoveryPages++
					env, tag, next := "dev-gcp", "dev-tag", true
					if vars.After != nil {
						env, tag, next = "prod-gcp", "prod-tag", false
					}
					// Substring matches must not become remediation targets.
					nodes := []any{reportTestNode("oracleverk-other", env, "unrelated", tag, &digest), reportTestNode("oracleverk", env, "registry/oracleverk", tag, &digest)}
					return map[string]any{"team": map[string]any{connection: reportTestConnection(nodes, next, new("workload-page-2"))}}, nil
				}
				if operation != "MCPVulnerabilityFindings" || vars.Environment != "dev-gcp" || vars.Applications != (kind == "application") {
					t.Fatalf("unexpected findings request: %s %+v", operation, vars)
				}
				findingPages++
				start := 0
				if vars.After != nil {
					var err error
					start, err = strconv.Atoi(*vars.After)
					if err != nil {
						t.Fatal(err)
					}
				}
				end := min(start+2, 3)
				nodes := []any{}
				for i := start; i < end; i++ {
					priority, severity, pkg := "MONITOR", "LOW", "pkg:golang/stdlib@v1.26.8"
					if i == 0 {
						severity = "CRITICAL"
					}
					if i == 2 {
						priority, severity, pkg = "HIGH", "HIGH", "pkg:golang/golang.org/x/net@v0.59.0"
					}
					nodes = append(nodes, map[string]any{
						"id": "finding-" + strconv.Itoa(i), "identifier": "CVE-shared", "package": pkg, "severity": severity, "description": "Advisory text",
						"vulnerabilityDetailsLink": "https://example.com/advisory",
						"riskAssessment":           map[string]any{"priority": priority, "epssScore": 0.02, "epssPercentile": 75.0},
						"remediation":              map[string]any{"fixVersion": nil, "latestVersion": nil},
						"suppression":              map[string]any{"state": "NOT_AFFECTED", "reason": "Existing documented assessment"},
					})
				}
				return reportTestFindingPage(kind, "registry/oracleverk", "dev-tag", &digest, nodes, end < 3, new(strconv.Itoa(end))), nil
			})

			report, err := getWorkloadVulnerabilityReport(t.Context(), client, GetWorkloadVulnerabilityReportInput{Team: "nais", Name: "oracleverk"})
			if err != nil {
				t.Fatal(err)
			}
			if !report.Complete || len(report.Images) != 1 || discoveryPages != 2 || findingPages != 2 {
				t.Fatalf("incomplete or duplicated report: %+v; requests: discovery=%d findings=%d", report, discoveryPages, findingPages)
			}
			image := report.Images[0]
			if len(image.Workloads) != 2 || len(image.Findings) != 3 || len(image.Warnings) != 0 {
				t.Fatalf("unexpected shared image report: %+v", image)
			}
			for _, workload := range image.Workloads {
				if workload.Kind != kind || workload.ImageTag == "" || workload.LatestDeployment == nil || *workload.LatestDeployment.CommitSha != "deployed-commit" {
					t.Errorf("missing workload/deployment context: %+v", workload)
				}
			}
			first := image.Findings[0]
			if first.Id != "finding-2" || first.Remediation.FixVersion != nil || first.Suppression == nil || first.Suppression.Reason != "Existing documented assessment" || first.RiskAssessment.EpssPercentile == nil || *first.RiskAssessment.EpssPercentile != 75 {
				t.Errorf("priority, unknown remediation or assessment was not preserved: %+v", first)
			}
			if image.Findings[0].Package == image.Findings[1].Package {
				t.Error("different package findings for the same advisory were collapsed")
			}
			if image.Findings[1].Id != "finding-0" {
				t.Error("severity was not used to break equal-priority ties")
			}
		})
	}
}

func TestWorkloadVulnerabilityReportImageGrouping(t *testing.T) {
	tests := []struct {
		name        string
		imageName   string
		tag         string
		digest      *string
		otherDigest *string
		wantImages  int
	}{
		{"different digests stay separate", "registry/oracleverk", "same", new("sha256:one"), new("sha256:two"), 2},
		{"same digest in different repositories stays separate", "other/oracleverk", "same", new("sha256:one"), new("sha256:one"), 2},
		{"missing digest groups identical tags", "registry/oracleverk", "same", nil, nil, 1},
		{"missing digest separates different tags", "registry/oracleverk", "other", nil, nil, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dev := reportTestNode("oracleverk", "dev-gcp", "registry/oracleverk", "same", tt.digest)
			prod := reportTestNode("oracleverk", "prod-gcp", tt.imageName, tt.tag, tt.otherDigest)
			calls := 0
			client := reportTestClient(func(_ context.Context, operation string, vars reportTestVariables) (any, error) {
				if operation == "MCPVulnerabilityWorkloads" {
					if !reflect.DeepEqual(vars.Environments, []string{"dev-gcp", "prod-gcp"}) || !vars.Jobs {
						t.Fatalf("selection was not forwarded: %+v", vars)
					}
					return map[string]any{"team": map[string]any{"jobs": reportTestConnection([]any{dev, prod}, false, nil)}}, nil
				}
				calls++
				if vars.Environment == "dev-gcp" {
					return reportTestFindingPage("job", "registry/oracleverk", "same", tt.digest, []any{}, false, nil), nil
				}
				return reportTestFindingPage("job", tt.imageName, tt.tag, tt.otherDigest, []any{}, false, nil), nil
			})
			report, err := getWorkloadVulnerabilityReport(t.Context(), client, GetWorkloadVulnerabilityReportInput{
				Team: "nais", Name: "oracleverk", Kind: "job", Environments: []string{"dev-gcp", "prod-gcp"},
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(report.Images) != tt.wantImages || calls != tt.wantImages {
				t.Fatalf("got %d images and %d fetches, want %d", len(report.Images), calls, tt.wantImages)
			}
			if tt.digest == nil && len(report.Images[0].Warnings) == 0 {
				t.Error("missing digest must warn about mutable image identity")
			}
		})
	}
}

func TestWorkloadVulnerabilityReportFailures(t *testing.T) {
	for _, failure := range []string{"no match", "ambiguous", "upstream findings error", "missing findings cursor", "repeated findings cursor", "image changed"} {
		t.Run(failure, func(t *testing.T) {
			digest := "sha256:original"
			client := reportTestClient(func(_ context.Context, operation string, vars reportTestVariables) (any, error) {
				if operation == "MCPVulnerabilityWorkloads" {
					connection := "applications"
					if vars.Jobs {
						connection = "jobs"
					}
					nodes := []any{}
					if failure == "ambiguous" || (vars.Jobs && failure != "no match") {
						nodes = append(nodes, reportTestNode("oracleverk", "dev-gcp", "registry/oracleverk", "tag", &digest))
					}
					return map[string]any{"team": map[string]any{connection: reportTestConnection(nodes, false, nil)}}, nil
				}
				switch failure {
				case "upstream findings error":
					return nil, errors.New("upstream unavailable")
				}
				pageDigest := digest
				if failure == "image changed" && vars.After != nil {
					pageDigest = "sha256:new"
				}
				next := failure == "missing findings cursor" || failure == "repeated findings cursor" || (failure == "image changed" && vars.After == nil)
				var cursor *string
				if failure != "missing findings cursor" {
					cursor = new("repeated")
				}
				return reportTestFindingPage("job", "registry/oracleverk", "tag", &pageDigest, []any{}, next, cursor), nil
			})
			report, err := getWorkloadVulnerabilityReport(t.Context(), client, GetWorkloadVulnerabilityReportInput{Team: "nais", Name: "oracleverk"})
			if err == nil || report.Complete || len(report.Images) != 0 {
				t.Fatalf("expected failure without a partial success; report=%+v err=%v", report, err)
			}
		})
	}
}

func TestWorkloadVulnerabilityReportFreshnessWarnings(t *testing.T) {
	for _, state := range []string{"missing summary", "missing timestamp", "older image", "missing SBOM"} {
		t.Run(state, func(t *testing.T) {
			digest := "sha256:image"
			node := reportTestNode("oracleverk", "dev-gcp", "registry/oracleverk", "tag", &digest)
			image := node["image"].(map[string]any)
			switch state {
			case "missing summary":
				image["vulnerabilitySummary"] = nil
			case "missing timestamp":
				image["vulnerabilitySummary"].(map[string]any)["lastUpdated"] = nil
			case "older image":
				image["vulnerabilitySummary"].(map[string]any)["staleImageTag"] = "previous-tag"
			case "missing SBOM":
				image["sbom"] = map[string]any{"status": "PENDING"}
			}
			client := reportTestClient(func(_ context.Context, operation string, _ reportTestVariables) (any, error) {
				if operation == "MCPVulnerabilityWorkloads" {
					return map[string]any{"team": map[string]any{"jobs": reportTestConnection([]any{node}, false, nil)}}, nil
				}
				return reportTestFindingPage("job", "registry/oracleverk", "tag", &digest, []any{}, false, nil), nil
			})
			report, err := getWorkloadVulnerabilityReport(t.Context(), client, GetWorkloadVulnerabilityReportInput{Team: "nais", Name: "oracleverk", Kind: "job"})
			if err != nil {
				t.Fatal(err)
			}
			if !report.Complete || len(report.Images[0].Warnings) != 1 || report.Images[0].Findings == nil {
				t.Fatalf("missing data looked like verified zero findings: %+v", report)
			}
		})
	}
}
