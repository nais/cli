package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/nais/cli/internal/mcp/client"
)

type vulnerabilityTestClient struct {
	client.Client
	fetch func(context.Context, client.WorkloadVulnerabilityRequest) (client.WorkloadVulnerabilityPage, error)
}

func (c *vulnerabilityTestClient) GetWorkloadVulnerabilities(ctx context.Context, request client.WorkloadVulnerabilityRequest) (client.WorkloadVulnerabilityPage, error) {
	return c.fetch(ctx, request)
}

type denyVulnerabilityRequests struct{}

func (denyVulnerabilityRequests) Allow() bool { return false }

func vulnerabilityArgs() GetWorkloadVulnerabilitiesInput {
	return GetWorkloadVulnerabilitiesInput{Team: "team-alpha", Workload: "app", Environment: "prod", WorkloadType: "application"}
}

func TestWorkloadVulnerabilitiesPages(t *testing.T) {
	for _, kind := range []string{"application", "job"} {
		t.Run(kind, func(t *testing.T) {
			handler := &toolContext{client: client.NewMockClient(client.ScenarioVulnerable), rateLimiter: allowAll{}}
			first := 1
			args := vulnerabilityArgs()
			args.WorkloadType, args.First = kind, &first
			var findings []client.WorkloadVulnerability
			var identity string
			for {
				result, err := handler.handleGetWorkloadVulnerabilities(context.Background(), mcp.CallToolRequest{}, args)
				if err != nil {
					t.Fatal(err)
				}
				if identity != "" && identity != result.SnapshotIdentity {
					t.Fatal("snapshot identity changed between pages")
				}
				identity = result.SnapshotIdentity
				findings = append(findings, result.Findings...)
				if result.ReturnedCount != 1 || len(result.CoverageIssues) != 0 {
					t.Fatalf("unexpected page coverage: %+v", result)
				}
				if !result.PageInfo.HasNextPage {
					break
				}
				args.After, args.ExpectedSnapshot = result.PageInfo.EndCursor, identity
			}
			if len(findings) != 3 || findings[0].Identifier != findings[1].Identifier || findings[0].Package == findings[1].Package {
				t.Fatal("did not retain the same CVE in different packages")
			}
			if findings[1].SuppressionState == nil || findings[1].Remediation.FixVersion != nil || findings[1].RiskAssessment.EPSSScore != nil {
				t.Fatal("suppression or unknown fix/risk was lost")
			}
		})
	}
}

func TestWorkloadVulnerabilitiesInput(t *testing.T) {
	tests := []struct {
		name string
		edit func(*GetWorkloadVulnerabilitiesInput)
		want string
	}{
		{"empty team", func(a *GetWorkloadVulnerabilitiesInput) { a.Team = "" }, "invalid_scope"},
		{"whitespace", func(a *GetWorkloadVulnerabilitiesInput) { a.Environment = " prod" }, "invalid_scope"},
		{"type", func(a *GetWorkloadVulnerabilitiesInput) { a.WorkloadType = "pod" }, "invalid_workload_type"},
		{"zero page", func(a *GetWorkloadVulnerabilitiesInput) { n := 0; a.First = &n }, "invalid_page_size"},
		{"large page", func(a *GetWorkloadVulnerabilitiesInput) { n := 51; a.First = &n }, "invalid_page_size"},
		{"cursor without snapshot", func(a *GetWorkloadVulnerabilitiesInput) { s := "cursor"; a.After = &s }, "invalid_cursor"},
		{"empty cursor", func(a *GetWorkloadVulnerabilitiesInput) { s := ""; a.After = &s; a.ExpectedSnapshot = "x" }, "invalid_cursor"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			handler := &toolContext{
				rateLimiter: allowAll{},
				client: &vulnerabilityTestClient{fetch: func(context.Context, client.WorkloadVulnerabilityRequest) (client.WorkloadVulnerabilityPage, error) {
					calls++
					return client.WorkloadVulnerabilityPage{}, nil
				}},
			}
			args := vulnerabilityArgs()
			test.edit(&args)
			_, err := handler.handleGetWorkloadVulnerabilities(context.Background(), mcp.CallToolRequest{}, args)
			if err == nil || !strings.Contains(err.Error(), test.want) || calls != 0 {
				t.Fatalf("err=%v calls=%d, want %s before API call", err, calls, test.want)
			}
		})
	}
}

func TestWorkloadVulnerabilitiesResponseGates(t *testing.T) {
	mock := client.NewMockClient(client.ScenarioVulnerable)
	args := vulnerabilityArgs()
	handler := &toolContext{client: mock, rateLimiter: allowAll{}}
	baseline, err := handler.handleGetWorkloadVulnerabilities(context.Background(), mcp.CallToolRequest{}, args)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		edit func(*client.WorkloadVulnerabilityPage)
		want string
	}{
		{"wrong workload", func(p *client.WorkloadVulnerabilityPage) { p.Workload.Name = "other" }, "scope_mismatch"},
		{"wrong environment", func(p *client.WorkloadVulnerabilityPage) { p.Workload.Environment = "dev" }, "scope_mismatch"},
		{"new image", func(p *client.WorkloadVulnerabilityPage) { p.Snapshot.ImageID = "new-image" }, "snapshot_changed"},
		{"new scan", func(p *client.WorkloadVulnerabilityPage) {
			updated := p.Snapshot.ScanUpdatedAt.AddDate(0, 0, 1)
			p.Snapshot.ScanUpdatedAt = &updated
		}, "snapshot_changed"},
		{"missing cursor", func(p *client.WorkloadVulnerabilityPage) { p.PageInfo.HasNextPage = true; p.PageInfo.EndCursor = nil }, "invalid_cursor"},
		{"repeated cursor", func(p *client.WorkloadVulnerabilityPage) { p.PageInfo.EndCursor = args.After }, "invalid_cursor"},
		{"duplicate row", func(p *client.WorkloadVulnerabilityPage) { p.Findings[1] = p.Findings[0] }, "invalid_response"},
	}
	cursor := "mock-1"
	args.After, args.ExpectedSnapshot = &cursor, baseline.SnapshotIdentity
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler.client = &vulnerabilityTestClient{fetch: func(ctx context.Context, request client.WorkloadVulnerabilityRequest) (client.WorkloadVulnerabilityPage, error) {
				page, err := mock.GetWorkloadVulnerabilities(ctx, request)
				test.edit(&page)
				return page, err
			}}
			_, err := handler.handleGetWorkloadVulnerabilities(context.Background(), mcp.CallToolRequest{}, args)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("err=%v, want %s", err, test.want)
			}
		})
	}
}

func TestWorkloadVulnerabilitiesCoverageAndPreview(t *testing.T) {
	mock := client.NewMockClient(client.ScenarioVulnerable)
	for _, test := range []struct {
		name string
		edit func(*client.WorkloadVulnerabilityPage)
		want string
	}{
		{"SBOM", func(p *client.WorkloadVulnerabilityPage) { p.Snapshot.SBOMStatus = "PROCESSING" }, "sbom_not_ready"},
		{"summary", func(p *client.WorkloadVulnerabilityPage) { p.Snapshot.Total = nil }, "summary_unavailable"},
		{"scan time", func(p *client.WorkloadVulnerabilityPage) { p.Snapshot.ScanUpdatedAt = nil }, "scan_time_unknown"},
	} {
		t.Run(test.name, func(t *testing.T) {
			handler := &toolContext{rateLimiter: allowAll{}, client: &vulnerabilityTestClient{fetch: func(ctx context.Context, r client.WorkloadVulnerabilityRequest) (client.WorkloadVulnerabilityPage, error) {
				page, err := mock.GetWorkloadVulnerabilities(ctx, r)
				test.edit(&page)
				page.Findings[0].Description = strings.Repeat("ø", 1001)
				return page, err
			}}}
			output, err := handler.handleGetWorkloadVulnerabilities(context.Background(), mcp.CallToolRequest{}, vulnerabilityArgs())
			if err != nil {
				t.Fatal(err)
			}
			if len(output.CoverageIssues) != 1 || output.CoverageIssues[0] != test.want || !output.Findings[0].DescriptionTruncated || len([]rune(output.Findings[0].Description)) != 1000 {
				t.Fatalf("unexpected output: %+v", output)
			}
			raw, err := json.Marshal(output)
			if err != nil {
				t.Fatal(err)
			}
			for _, forbidden := range []string{`"priority"`, `"reason"`, `"manifest"`, `"email"`, `"token"`} {
				if strings.Contains(string(raw), forbidden) {
					t.Fatalf("output contains %s", forbidden)
				}
			}
		})
	}
}

func TestWorkloadVulnerabilitiesSummaryExcludesSuppressed(t *testing.T) {
	for _, open := range []int{2, 0} {
		t.Run(fmt.Sprintf("open-%d", open), func(t *testing.T) {
			handler := &toolContext{rateLimiter: allowAll{}, client: &vulnerabilityTestClient{
				fetch: func(ctx context.Context, request client.WorkloadVulnerabilityRequest) (client.WorkloadVulnerabilityPage, error) {
					page, err := client.NewMockClient(client.ScenarioVulnerable).GetWorkloadVulnerabilities(ctx, request)
					if err != nil {
						return page, err
					}
					seed := page.Findings[0]
					page.Findings = nil
					state := "NOT_AFFECTED"
					for i := range 7 {
						row := seed
						row.ID = fmt.Sprintf("finding-%d", i)
						if i >= open {
							row.SuppressionState = &state
						}
						page.Findings = append(page.Findings, row)
					}
					page.Snapshot.Total = &open
					page.PageInfo = client.VulnerabilityPageInfo{}
					return page, nil
				},
			}}
			output, err := handler.handleGetWorkloadVulnerabilities(context.Background(), mcp.CallToolRequest{}, vulnerabilityArgs())
			if err != nil {
				t.Fatal(err)
			}
			if len(output.Findings) != 7 || *output.Snapshot.Total != open || len(output.CoverageIssues) != 0 {
				t.Fatalf("suppressed findings rejected or summary changed: %+v", output)
			}
		})
	}
}

func TestWorkloadVulnerabilitiesErrors(t *testing.T) {
	handler := &toolContext{rateLimiter: denyVulnerabilityRequests{}}
	if _, err := handler.handleGetWorkloadVulnerabilities(context.Background(), mcp.CallToolRequest{}, vulnerabilityArgs()); err == nil || !strings.Contains(err.Error(), "rate limit") {
		t.Fatalf("rate limit err=%v", err)
	}
	handler.rateLimiter = allowAll{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := handler.handleGetWorkloadVulnerabilities(ctx, mcp.CallToolRequest{}, vulnerabilityArgs()); err != context.Canceled {
		t.Fatalf("cancellation err=%v", err)
	}
	handler.client = &vulnerabilityTestClient{fetch: func(context.Context, client.WorkloadVulnerabilityRequest) (client.WorkloadVulnerabilityPage, error) {
		return client.WorkloadVulnerabilityPage{}, fmt.Errorf("unauthenticated")
	}}
	if _, err := handler.handleGetWorkloadVulnerabilities(context.Background(), mcp.CallToolRequest{}, vulnerabilityArgs()); err == nil || !strings.Contains(err.Error(), "unauthenticated") {
		t.Fatalf("authentication err=%v", err)
	}
}
