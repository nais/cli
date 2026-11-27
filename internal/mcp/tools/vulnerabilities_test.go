package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/Khan/genqlient/graphql"
	"github.com/nais/cli/internal/vulnerability"
)

type reportTestVariables struct {
	Team         string   `json:"team"`
	Name         string   `json:"name"`
	Environment  string   `json:"environment"`
	Environments []string `json:"environments"`
	After        *string  `json:"after"`
	Applications bool     `json:"applications"`
	Jobs         bool     `json:"jobs"`
	Overview     bool     `json:"overview"`
	Filter       struct {
		Suppressed string `json:"suppressed"`
	} `json:"filter"`
	Order struct {
		Field     string `json:"field"`
		Direction string `json:"direction"`
	} `json:"order"`
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
	result, requestErr := c(ctx, req.OpName, variables)
	if result == nil {
		return requestErr
	}
	data, err = json.Marshal(result)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, resp.Data); err != nil {
		return err
	}
	return requestErr
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

func findingReportTestNode(kind, environment, imageName, tag string, digest *string) map[string]any {
	node := reportTestNode("oracleverk", environment, imageName, tag, digest)
	node["id"], node["team"] = kind+"-"+environment, map[string]any{"slug": "nais"}
	node["__typename"] = "Application"
	if kind == "job" {
		node["__typename"] = "Job"
	}
	status := "SCANNED"
	if digest == nil {
		status = "PENDING"
	}
	node["vulnerabilityOverview"] = map[string]any{
		"findingCount": 3, "uniqueVulnerabilityCount": 2, "knownExploitedVulnerabilityCount": 1,
		"countsBySeverity": map[string]any{"critical": 1, "high": 1, "medium": 0, "low": 0, "unknown": 0},
		"coverage":         map[string]any{"status": status, "imageName": imageName, "imageDigest": digest, "scannedAt": "2026-11-16T09:00:00Z", "error": nil},
	}
	node["vulnerabilityCoverage"] = node["vulnerabilityOverview"].(map[string]any)["coverage"]
	return node
}

func findingReportTestFinding(workload map[string]any, identifier, purl string, fixed, reason *string) map[string]any {
	var suppression any
	if reason != nil {
		suppression = map[string]any{"state": "NOT_AFFECTED", "reason": *reason}
	}
	image := workload["image"].(map[string]any)
	return map[string]any{
		"id": fmt.Sprintf("%s-%s-%s", workload["id"], purl, identifier), "vulnerabilityIdentifier": identifier,
		"purl": purl, "packageKey": strings.Split(purl, "@")[0], "installedVersion": "v1.26.8", "fixedVersion": fixed,
		"severity": "HIGH", "severitySource": "NVD", "dataSource": "OSV", "knownExploited": true,
		"epss": map[string]any{"score": 0.02, "date": "2026-11-16"},
		"vulnerability": map[string]any{
			"id": identifier, "identifier": identifier, "aliases": []string{"GHSA-alias"}, "references": []string{"https://example.invalid/advisory"},
			"title": "Advisory", "description": "Untrusted advisory text", "knownExploited": true, "epss": nil,
		},
		"suppression": suppression, "image": map[string]any{"name": image["name"], "digest": image["digest"]}, "workload": workload,
	}
}

func findingReportTestPage(workload map[string]any, nodes []any, next bool, cursor *string) map[string]any {
	parent := make(map[string]any)
	for key, value := range workload {
		parent[key] = value
	}
	parent["vulnerabilityFindings"] = reportTestConnection(nodes, next, cursor)
	return map[string]any{"team": map[string]any{"environment": map[string]any{strings.ToLower(parent["__typename"].(string)): parent}}}
}

func TestWorkloadVulnerabilityReport(t *testing.T) {
	digest := new("sha256:shared")
	dev := findingReportTestNode("application", "dev-gcp", "registry/oracleverk", "dev-tag", digest)
	prod := findingReportTestNode("application", "prod-gcp", "registry/oracleverk", "prod-tag", digest)
	dev["vulnerabilityOverview"].(map[string]any)["coverage"].(map[string]any)["error"] = "Rescan failed"
	discoveryPages, findingPages := 0, 0
	stdlib, net := "pkg:golang/stdlib@v1.26.8", "pkg:golang/golang.org/x/net@v0.59.0"
	client := reportTestClient(func(_ context.Context, operation string, vars reportTestVariables) (any, error) {
		if vars.Team != "nais" || vars.Name != "oracleverk" || vars.Filter.Suppressed != "INCLUDE" {
			t.Fatalf("unexpected target or suppression filter: %+v", vars)
		}
		if operation == "MCPVulnerabilityTargets" {
			discoveryPages++
			if !vars.Applications || vars.Jobs || !vars.Overview {
				t.Fatalf("incorrect workload kind: %+v", vars)
			}
			node, next := dev, true
			if vars.After != nil {
				node, next = prod, false
			}
			return map[string]any{"team": map[string]any{"applications": reportTestConnection([]any{node}, next, new("page-2"))}}, nil
		}
		if operation != "WorkloadVulnerabilityFindings" || vars.Order.Field != "SEVERITY" || vars.Order.Direction != "DESC" {
			t.Fatalf("unexpected findings request: %s %+v", operation, vars)
		}
		findingPages++
		if vars.Environment == "prod-gcp" {
			return findingReportTestPage(prod, []any{
				findingReportTestFinding(prod, "CVE-1", stdlib, new("v1.26.11"), nil),
				findingReportTestFinding(prod, "CVE-2", stdlib, nil, new("Prod assessment")),
			}, false, nil), nil
		}
		if vars.After == nil {
			return findingReportTestPage(dev, []any{
				findingReportTestFinding(dev, "CVE-1", stdlib, new("v1.26.9"), new("Dev assessment")),
			}, true, new("page-2")), nil
		}
		unknown := findingReportTestFinding(dev, "CVE-1", net, nil, nil)
		unknown["vulnerability"], unknown["epss"] = nil, nil
		return findingReportTestPage(dev, []any{findingReportTestFinding(dev, "CVE-2", stdlib, nil, nil), unknown}, false, nil), nil
	})
	report, err := getWorkloadVulnerabilityReport(t.Context(), client, GetWorkloadVulnerabilityReportInput{
		VulnerabilityWorkloadTarget: VulnerabilityWorkloadTarget{Team: "nais", Name: "oracleverk", Kind: "application"}, IncludeDetails: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !report.Complete || !report.IncludeSuppressed || len(report.Images) != 1 || discoveryPages != 4 || findingPages != 3 {
		t.Fatalf("incomplete or duplicated report: %+v; requests: discovery=%d findings=%d", report, discoveryPages, findingPages)
	}
	image := report.Images[0]
	if len(image.Workloads) != 2 || len(image.Packages) != 2 || len(image.Findings) != 5 {
		t.Fatalf("missing workload/package/finding context: %+v", image)
	}
	netPackage, stdlibPackage := image.Packages[0], image.Packages[1]
	if stdlibPackage.Findings != 2 || stdlibPackage.UniqueVulnerabilityCount != 2 || stdlibPackage.BySeverity != (vulnerability.SeverityCounts{High: 2}) || stdlibPackage.Suppressed != 2 || stdlibPackage.WithoutKnownFix != 1 ||
		!reflect.DeepEqual(stdlibPackage.FixVersions, []string{"v1.26.11", "v1.26.9"}) || netPackage.Findings != 1 {
		t.Errorf("shared-image findings or fix versions counted incorrectly: %+v", image.Packages)
	}
	details := map[string]FindingReportDetails{}
	for _, finding := range image.Findings {
		details[finding.Environment+"/"+finding.PURL+"/"+finding.Identifier] = finding
	}
	devAssessment := details["dev-gcp/"+stdlib+"/CVE-1"]
	prodAssessment := details["prod-gcp/"+stdlib+"/CVE-2"]
	unknown := details["dev-gcp/"+net+"/CVE-1"]
	if devAssessment.Suppression == nil || devAssessment.Suppression.Reason != "Dev assessment" ||
		details["prod-gcp/"+stdlib+"/CVE-1"].Suppression != nil || prodAssessment.Suppression == nil ||
		prodAssessment.Suppression.Reason != "Prod assessment" || devAssessment.EPSS == nil || devAssessment.EPSS.Date != "2026-11-16" ||
		unknown.Advisory != nil || unknown.EPSS != nil {
		t.Errorf("nullable details or workload-specific assessments lost: %+v", image.Findings)
	}
	if image.Workloads[0].Coverage.Status != "SCANNED" || image.Workloads[0].Coverage.Error == nil || len(image.Warnings) != 1 ||
		image.Workloads[0].Overview.UniqueVulnerabilityCount != 2 || image.Workloads[0].LatestDeployment == nil {
		t.Errorf("rescan error or workload overview/deployment lost: %+v", image)
	}
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	for _, removed := range []string{`"by_priority"`, `"reported_latest_versions"`, `"riskAssessment"`, `"updatedBy"`, `"scan_state"`, `"sbom"`} {
		if strings.Contains(string(data), removed) {
			t.Errorf("report still exposes removed field %s", removed)
		}
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
			dev := findingReportTestNode("job", "dev-gcp", "registry/oracleverk", "same", tt.digest)
			prod := findingReportTestNode("job", "prod-gcp", tt.imageName, tt.tag, tt.otherDigest)
			for _, node := range []map[string]any{dev, prod} {
				overview := node["vulnerabilityOverview"].(map[string]any)
				overview["findingCount"], overview["uniqueVulnerabilityCount"], overview["knownExploitedVulnerabilityCount"] = 0, 0, 0
				overview["countsBySeverity"] = map[string]int{}
			}
			calls := 0
			client := reportTestClient(func(_ context.Context, operation string, vars reportTestVariables) (any, error) {
				if operation == "MCPVulnerabilityTargets" {
					if !reflect.DeepEqual(vars.Environments, []string{"dev-gcp", "prod-gcp"}) || !vars.Jobs {
						t.Fatalf("selection was not forwarded: %+v", vars)
					}
					return map[string]any{"team": map[string]any{"jobs": reportTestConnection([]any{dev, prod}, false, nil)}}, nil
				}
				calls++
				node := dev
				if vars.Environment == "prod-gcp" {
					node = prod
				}
				return findingReportTestPage(node, []any{}, false, nil), nil
			})
			report, err := getWorkloadVulnerabilityReport(t.Context(), client, GetWorkloadVulnerabilityReportInput{
				VulnerabilityWorkloadTarget: VulnerabilityWorkloadTarget{
					Team: "nais", Name: "oracleverk", Kind: "job", Environments: []string{"dev-gcp", "prod-gcp"},
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(report.Images) != tt.wantImages || calls != 2 || len(report.Images[0].Packages) != 0 || len(report.Images[0].Findings) != 0 {
				t.Fatalf("got %d images and %d workload fetches, want %d images and 2 fetches", len(report.Images), calls, tt.wantImages)
			}
			if tt.digest == nil && len(report.Images[0].Warnings) == 0 {
				t.Error("missing digest must warn about mutable image identity")
			}
		})
	}
}

func TestWorkloadVulnerabilityReportCoverageFallbacks(t *testing.T) {
	for _, tt := range []struct {
		name         string
		digest       *string
		coverage     string
		wantFindings int
		wantWarning  string
	}{
		{"tag-only deployment that was scanned", nil, "registry/oracleverk", 1, "Image digest unavailable"},
		{"coverage for another image", new("sha256:deployed"), "registry/other", 0, "coverage describes a different image"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			node := findingReportTestNode("job", "dev-gcp", "registry/oracleverk", "tag", tt.digest)
			node["vulnerabilityCoverage"] = map[string]any{
				"status": "SCANNED", "imageName": tt.coverage, "imageDigest": "sha256:resolved", "scannedAt": "2026-11-16T09:00:00Z", "error": nil,
			}
			findingCalls := 0
			client := reportTestClient(func(_ context.Context, operation string, _ reportTestVariables) (any, error) {
				if operation == "MCPVulnerabilityTargets" {
					return map[string]any{"team": map[string]any{"jobs": reportTestConnection([]any{node}, false, nil)}}, nil
				}
				findingCalls++
				finding := findingReportTestFinding(node, "CVE-1", "pkg:golang/stdlib@v1.26.8", nil, nil)
				finding["image"] = map[string]any{"name": "registry/oracleverk", "digest": "sha256:resolved"}
				return findingReportTestPage(node, []any{finding}, false, nil), nil
			})
			report, err := getWorkloadVulnerabilityReport(t.Context(), client, GetWorkloadVulnerabilityReportInput{
				VulnerabilityWorkloadTarget: VulnerabilityWorkloadTarget{Team: "nais", Name: "oracleverk", Kind: "job"}, IncludeDetails: true,
			})
			if err != nil {
				t.Fatal(err)
			}
			image := report.Images[0]
			if len(image.Findings) != tt.wantFindings || findingCalls != tt.wantFindings || !strings.Contains(strings.Join(image.Warnings, " "), tt.wantWarning) {
				t.Fatalf("findings=%d calls=%d warnings=%v", len(image.Findings), findingCalls, image.Warnings)
			}
		})
	}
}

func TestWorkloadVulnerabilityReportFailures(t *testing.T) {
	upstream := errors.New("upstream unavailable")
	for _, failure := range []string{"no match", "ambiguous", "missing environment", "later findings error", "image changed on empty page", "finding scope mismatch", "final image change", "missing overview"} {
		t.Run(failure, func(t *testing.T) {
			dev := findingReportTestNode("job", "dev-gcp", "registry/oracleverk", "tag", new("sha256:original"))
			discoveries := 0
			client := reportTestClient(func(_ context.Context, operation string, vars reportTestVariables) (any, error) {
				if operation == "MCPVulnerabilityTargets" {
					discoveries++
					connection := "applications"
					nodes := []any{}
					if vars.Jobs {
						connection = "jobs"
					}
					node := dev
					if failure == "ambiguous" && vars.Applications {
						node = findingReportTestNode("application", "dev-gcp", "registry/oracleverk", "tag", new("sha256:original"))
					}
					if failure == "final image change" && discoveries > 2 {
						node = findingReportTestNode("job", "dev-gcp", "registry/oracleverk", "tag", new("sha256:new"))
					}
					if failure == "missing overview" {
						node["vulnerabilityOverview"] = nil
					}
					if failure != "no match" && (vars.Jobs || failure == "ambiguous") {
						nodes = []any{node}
					}
					return map[string]any{"team": map[string]any{connection: reportTestConnection(nodes, false, nil)}}, nil
				}
				if failure == "later findings error" && vars.After != nil {
					return findingReportTestPage(dev, []any{}, false, nil), upstream
				}
				if failure == "image changed on empty page" && vars.After != nil {
					changed := findingReportTestNode("job", "dev-gcp", "registry/oracleverk", "tag", new("sha256:new"))
					return findingReportTestPage(changed, []any{}, false, nil), nil
				}
				finding := findingReportTestFinding(dev, "CVE-1", "pkg:golang/stdlib@v1.26.8", nil, nil)
				if failure == "finding scope mismatch" {
					finding["workload"] = findingReportTestNode("job", "other-gcp", "registry/oracleverk", "tag", new("sha256:original"))
				}
				next := vars.After == nil && (failure == "later findings error" || failure == "image changed on empty page")
				return findingReportTestPage(dev, []any{finding}, next, new("page-2")), nil
			})
			args := GetWorkloadVulnerabilityReportInput{VulnerabilityWorkloadTarget: VulnerabilityWorkloadTarget{Team: "nais", Name: "oracleverk"}}
			if failure == "missing environment" {
				args.Environments = []string{"prod-gcp"}
			}
			report, err := getWorkloadVulnerabilityReport(t.Context(), client, args)
			if err == nil || report.Complete || len(report.Images) != 0 {
				t.Fatalf("expected failure without partial success for %s; report=%+v err=%v", failure, report, err)
			}
			if failure == "later findings error" && !errors.Is(err, upstream) {
				t.Errorf("upstream failure lost: %v", err)
			}
		})
	}
}

func TestWorkloadVulnerabilityStatus(t *testing.T) {
	digest, repository, scanned := new("sha256:image"), new("registry/oracleverk"), new("2026-11-16T09:00:00Z")
	tests := []struct {
		name           string
		status         string
		digest         *string
		coverageName   *string
		coverageDigest *string
		scannedAt      *string
		scanError      *string
		wantState      string
	}{
		{"scanned", "SCANNED", digest, repository, digest, scanned, nil, "scan_ready"},
		{"pending", "PENDING", digest, repository, digest, nil, nil, "scan_pending"},
		{"failed", "FAILED", digest, repository, digest, nil, nil, "scan_unavailable"},
		{"failed rescan", "SCANNED", digest, repository, digest, scanned, new("Rescan failed"), "scan_error"},
		{"missing timestamp", "SCANNED", digest, repository, digest, nil, nil, "scan_unavailable"},
		{"zero timestamp", "SCANNED", digest, repository, digest, new("0001-01-01T00:00:00Z"), nil, "scan_unavailable"},
		{"missing digest", "PENDING", nil, repository, nil, nil, nil, "digest_unavailable"},
		{"partial rollout", "SCANNED", new("sha256:old"), repository, new("sha256:old"), scanned, nil, "not_deployed"},
		{"wrong coverage digest", "SCANNED", digest, repository, new("sha256:old"), scanned, nil, "scan_unavailable"},
		{"wrong coverage repository", "SCANNED", digest, new("other/repository"), digest, scanned, nil, "scan_unavailable"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := findingReportTestNode("job", "dev-gcp", "registry/oracleverk", "tag", digest)
			prod := findingReportTestNode("job", "prod-gcp", "registry/oracleverk", "tag", tt.digest)
			delete(node, "vulnerabilityOverview")
			delete(prod, "vulnerabilityOverview")
			prod["vulnerabilityCoverage"] = map[string]any{
				"status": tt.status, "imageName": tt.coverageName, "imageDigest": tt.coverageDigest,
				"scannedAt": tt.scannedAt, "error": tt.scanError,
			}
			client := reportTestClient(func(_ context.Context, operation string, vars reportTestVariables) (any, error) {
				if operation != "MCPVulnerabilityTargets" || vars.Overview || vars.Filter.Suppressed != "" {
					t.Fatal("status must only read workload identity and coverage, never findings or counts")
				}
				return map[string]any{"team": map[string]any{"jobs": reportTestConnection([]any{node, prod}, false, nil)}}, nil
			})
			report, err := getWorkloadVulnerabilityStatus(t.Context(), client, GetWorkloadVulnerabilityStatusInput{
				VulnerabilityWorkloadTarget: VulnerabilityWorkloadTarget{
					Team: "nais", Name: "oracleverk", Kind: "job", Environments: []string{"dev-gcp", "prod-gcp"},
				},
				ExpectedDigest: *digest,
			})
			if err != nil {
				t.Fatal(err)
			}
			if report.Verified != (tt.wantState == "scan_ready") {
				t.Fatalf("incorrect verification result for %q: %+v", tt.name, report)
			}
			found := false
			for _, group := range report.Images {
				for _, workload := range group.Workloads {
					if workload.Environment != "prod-gcp" {
						continue
					}
					found = true
					if group.State != tt.wantState || !reflect.DeepEqual(workload.Coverage.Error, tt.scanError) {
						t.Errorf("got state %q and error %v, want %q and %v", group.State, workload.Coverage.Error, tt.wantState, tt.scanError)
					}
				}
			}
			if !found {
				t.Error("requested production workload missing")
			}
		})
	}
}

func TestVulnerabilityStatusUsesLatestCoverage(t *testing.T) {
	dev := findingReportTestNode("application", "dev-gcp", "registry/oracleverk", "tag", new("sha256:image"))
	prod := findingReportTestNode("application", "prod-gcp", "registry/oracleverk", "tag", new("sha256:image"))
	delete(dev, "vulnerabilityOverview")
	delete(prod, "vulnerabilityOverview")
	passes := 0
	client := reportTestClient(func(_ context.Context, operation string, vars reportTestVariables) (any, error) {
		if operation != "MCPVulnerabilityTargets" || vars.Overview || !vars.Applications || vars.Jobs {
			t.Fatalf("unexpected coverage read: %s %+v", operation, vars)
		}
		if vars.After == nil {
			passes++
			return map[string]any{"team": map[string]any{"applications": reportTestConnection([]any{dev}, true, new("prod-page"))}}, nil
		}
		if passes == 2 {
			prod["vulnerabilityCoverage"].(map[string]any)["error"] = "Latest rescan failed"
		}
		return map[string]any{"team": map[string]any{"applications": reportTestConnection([]any{prod}, false, nil)}}, nil
	})
	status, err := getWorkloadVulnerabilityStatus(t.Context(), client, GetWorkloadVulnerabilityStatusInput{
		VulnerabilityWorkloadTarget: VulnerabilityWorkloadTarget{Team: "nais", Name: "oracleverk", Kind: "application"},
		ExpectedDigest:              "sha256:image",
	})
	if err != nil {
		t.Fatal(err)
	}
	if status.Verified || len(status.Images) != 1 || status.Images[0].State != "scan_error" || passes != 2 ||
		status.Images[0].Workloads[1].Coverage.Error == nil {
		t.Fatalf("latest paginated coverage error must prevent verification: %+v", status)
	}
}

func TestVulnerabilityStatusFailsWithoutPartialSuccess(t *testing.T) {
	upstream := errors.New("upstream unavailable")
	for _, failure := range []string{"image changed", "partial API error", "missing coverage"} {
		t.Run(failure, func(t *testing.T) {
			calls := 0
			client := reportTestClient(func(_ context.Context, _ string, _ reportTestVariables) (any, error) {
				calls++
				node := findingReportTestNode("job", "dev-gcp", "registry/oracleverk", "tag", new("sha256:image"))
				if calls == 2 && failure == "image changed" {
					node["image"].(map[string]any)["digest"] = "sha256:new"
				}
				if failure == "missing coverage" {
					node["vulnerabilityCoverage"] = nil
				}
				response := map[string]any{"team": map[string]any{"jobs": reportTestConnection([]any{node}, false, nil)}}
				if calls == 2 && failure == "partial API error" {
					return response, upstream
				}
				return response, nil
			})
			status, err := getWorkloadVulnerabilityStatus(t.Context(), client, GetWorkloadVulnerabilityStatusInput{
				VulnerabilityWorkloadTarget: VulnerabilityWorkloadTarget{Team: "nais", Name: "oracleverk", Kind: "job"},
				ExpectedDigest:              "sha256:image",
			})
			if err == nil || status.Verified || len(status.Images) != 0 {
				t.Fatalf("expected error without partial status for %s: status=%+v err=%v", failure, status, err)
			}
			if failure == "partial API error" && !errors.Is(err, upstream) {
				t.Errorf("API error lost: %v", err)
			}
		})
	}
}
