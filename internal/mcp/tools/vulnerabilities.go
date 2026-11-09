package tools

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nais/cli/internal/mcp/client"
)

type GetWorkloadVulnerabilitiesInput struct {
	Team             string  `json:"team" jsonschema:"required"`
	Workload         string  `json:"workload" jsonschema:"required"`
	Environment      string  `json:"environment" jsonschema:"required"`
	WorkloadType     string  `json:"workload_type" jsonschema:"required,enum=application,enum=job"`
	First            *int    `json:"first,omitempty" jsonschema_description:"Findings per page; default 10, maximum 50."`
	After            *string `json:"after,omitempty"`
	ExpectedSnapshot string  `json:"expected_snapshot,omitempty" jsonschema_description:"Snapshot identity from the first page; required with after."`
}

type GetWorkloadVulnerabilitiesOutput struct {
	client.WorkloadVulnerabilityPage
	FetchedAt        time.Time `json:"fetched_at"`
	SnapshotIdentity string    `json:"snapshot_identity" jsonschema_description:"Observed consistency fingerprint, not an atomic snapshot or authorization token."`
	ReturnedCount    int       `json:"returned_count"`
	CoverageIssues   []string  `json:"coverage_issues" jsonschema_description:"Empty means no observed coverage gaps on this page, not that all findings were fetched."`
}

func registerVulnerabilityTools(s *server.MCPServer, ctx *toolContext) {
	s.AddTool(mcp.NewTool(
		"get_workload_vulnerabilities",
		mcp.WithDescription("Get one bounded page of vulnerabilities and exposure context for an exact Nais application or job. Includes suppressed findings, reported risk and remediation, image/SBOM/summary metadata. Summary update time is not scan time; scan time remains unknown. Follow page_info until complete, passing snapshot_identity as expected_snapshot with after. Never mix snapshots; snapshot_changed requires a new collection. No code analysis, fixes, secrets or atomic-snapshot guarantee. EPSS score is 0-1; percentile is 0-100. Descriptions may be truncated previews; consult details_link for the full advisory."),
		mcp.WithInputSchema[GetWorkloadVulnerabilitiesInput](),
		mcp.WithOutputSchema[GetWorkloadVulnerabilitiesOutput](),
	), mcp.NewStructuredToolHandler(ctx.handleGetWorkloadVulnerabilities))
}

func (t *toolContext) handleGetWorkloadVulnerabilities(ctx context.Context, _ mcp.CallToolRequest, args GetWorkloadVulnerabilitiesInput) (GetWorkloadVulnerabilitiesOutput, error) {
	var output GetWorkloadVulnerabilitiesOutput
	if !t.rateLimiter.Allow() {
		return output, fmt.Errorf("rate limit exceeded, please try again later")
	}
	if err := ctx.Err(); err != nil {
		return output, err
	}
	for _, value := range []string{args.Team, args.Workload, args.Environment} {
		if strings.TrimSpace(value) == "" || value != strings.TrimSpace(value) {
			return output, fmt.Errorf("invalid_scope: team, workload and environment must be nonempty without surrounding whitespace")
		}
	}
	if args.WorkloadType != "application" && args.WorkloadType != "job" {
		return output, fmt.Errorf("invalid_workload_type: use application or job")
	}
	first := 10
	if args.First != nil {
		first = *args.First
	}
	if first < 1 || first > 50 {
		return output, fmt.Errorf("invalid_page_size: first must be between 1 and 50")
	}
	if args.After != nil && (*args.After == "" || args.ExpectedSnapshot == "") {
		return output, fmt.Errorf("invalid_cursor: after requires a nonempty cursor and expected_snapshot")
	}
	page, err := t.client.GetWorkloadVulnerabilities(ctx, client.WorkloadVulnerabilityRequest{
		Team: args.Team, Workload: args.Workload, Environment: args.Environment, Type: args.WorkloadType, First: first, After: args.After,
	})
	if err != nil {
		return output, fmt.Errorf("vulnerability query failed: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return output, err
	}
	if page.Workload.Team != args.Team || page.Workload.Name != args.Workload || page.Workload.Environment != args.Environment || page.Workload.Type != args.WorkloadType || page.Workload.ID == "" {
		return output, fmt.Errorf("scope_mismatch: response does not match the requested workload")
	}
	if page.Snapshot.ImageID == "" || page.Snapshot.ImageName == "" || page.Snapshot.SBOMStatus == "" || len(page.Findings) > first {
		return output, fmt.Errorf("invalid_response: missing snapshot metadata or oversized page")
	}
	identityData, err := json.Marshal(struct {
		Workload client.WorkloadIdentity
		Snapshot client.VulnerabilitySnapshot
	}{page.Workload, page.Snapshot})
	if err != nil {
		return output, fmt.Errorf("snapshot encoding failed: %w", err)
	}
	hash := sha256.Sum256(identityData)
	identity := hex.EncodeToString(hash[:])
	if args.ExpectedSnapshot != "" && args.ExpectedSnapshot != identity {
		return output, fmt.Errorf("snapshot_changed: discard pending aggregation and collect a new snapshot")
	}
	if page.PageInfo.HasNextPage && (page.PageInfo.EndCursor == nil || *page.PageInfo.EndCursor == "" || len(page.Findings) == 0) ||
		args.After != nil && page.PageInfo.EndCursor != nil && *page.PageInfo.EndCursor == *args.After {
		return output, fmt.Errorf("invalid_cursor: next page cursor is missing or did not advance")
	}
	page.PageInfo.RequestedAfter = args.After
	page.Findings = slices.Clone(page.Findings)
	seen := make(map[string]bool, len(page.Findings))
	for i := range page.Findings {
		finding := &page.Findings[i]
		if finding.ID == "" || finding.Identifier == "" || finding.Package == "" || finding.Severity == "" || seen[finding.ID] {
			return output, fmt.Errorf("invalid_response: incomplete or duplicate finding")
		}
		seen[finding.ID] = true
		runes := []rune(finding.Description)
		if len(runes) > 1000 {
			finding.Description = string(runes[:1000])
			finding.DescriptionTruncated = true
		}
	}
	issues := []string{}
	if page.Snapshot.SBOMStatus != "READY" {
		issues = append(issues, "sbom_not_ready")
	}
	if page.Snapshot.Total == nil {
		issues = append(issues, "summary_unavailable")
	} else if *page.Snapshot.Total < 0 {
		return output, fmt.Errorf("invalid_response: summary total cannot be negative")
	}
	issues = append(issues, "scan_time_unknown")
	return GetWorkloadVulnerabilitiesOutput{
		WorkloadVulnerabilityPage: page, FetchedAt: time.Now().UTC(), SnapshotIdentity: identity,
		ReturnedCount: len(page.Findings), CoverageIssues: issues,
	}, nil
}
