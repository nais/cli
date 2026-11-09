package client

import (
	"context"
	"fmt"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/nais/cli/internal/naisapi/gql"
)

type WorkloadVulnerabilityRequest struct {
	Team        string
	Workload    string
	Environment string
	Type        string
	First       int
	After       *string
}

type WorkloadVulnerabilityPage struct {
	Workload WorkloadIdentity        `json:"workload"`
	Context  VulnerabilityContext    `json:"context"`
	Snapshot VulnerabilitySnapshot   `json:"snapshot"`
	Findings []WorkloadVulnerability `json:"findings"`
	PageInfo VulnerabilityPageInfo   `json:"page_info"`
}

type WorkloadIdentity struct {
	ID          string `json:"id"`
	Team        string `json:"team"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Environment string `json:"environment"`
}

type VulnerabilityContext struct {
	Ingresses        []VulnerabilityIngress     `json:"ingresses"`
	AuthIntegrations []string                   `json:"auth_integrations"`
	InboundRules     []VulnerabilityInboundRule `json:"inbound_rules"`
}

type VulnerabilityIngress struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

type VulnerabilityInboundRule struct {
	Workload string `json:"workload"`
	Team     string `json:"team"`
	Mutual   bool   `json:"mutual"`
}

type VulnerabilitySnapshot struct {
	ImageID          string     `json:"image_id"`
	ImageName        string     `json:"image_name"`
	ImageTag         string     `json:"image_tag"`
	ImageDigest      *string    `json:"image_digest"`
	SBOMStatus       string     `json:"sbom_status"`
	Total            *int       `json:"total" jsonschema_description:"Backend summary's unsuppressed finding count; not proof of current-image coverage. Suppressed findings remain listed."`
	SummaryUpdatedAt *time.Time `json:"summary_updated_at" jsonschema_description:"Last summary update, not a scan timestamp or proof of current-image scan freshness."`
}

type WorkloadVulnerability struct {
	ID                   string                   `json:"id"`
	Identifier           string                   `json:"identifier"`
	Package              string                   `json:"package"`
	Severity             string                   `json:"severity"`
	Description          string                   `json:"description"`
	DescriptionTruncated bool                     `json:"description_truncated"`
	RiskAssessment       VulnerabilityRisk        `json:"risk_assessment"`
	Remediation          VulnerabilityRemediation `json:"remediation"`
	SuppressionState     *string                  `json:"suppression_state"`
	DetailsLink          string                   `json:"details_link"`
}

type VulnerabilityRisk struct {
	CVSSScore          *float64 `json:"cvss_score"`
	EPSSScore          *float64 `json:"epss_score" jsonschema_description:"Probability of exploitation, 0-1; null means unknown."`
	EPSSPercentile     *float64 `json:"epss_percentile" jsonschema_description:"EPSS percentile as a percentage, 0-100; null means unknown."`
	HasKEVEntry        bool     `json:"has_kev_entry"`
	KnownRansomwareUse bool     `json:"known_ransomware_use"`
}

type VulnerabilityRemediation struct {
	FixVersion    *string `json:"fix_version"`
	LatestVersion *string `json:"latest_version"`
}

type VulnerabilityPageInfo struct {
	HasNextPage    bool    `json:"has_next_page"`
	EndCursor      *string `json:"end_cursor"`
	RequestedAfter *string `json:"requested_after"`
}

func fetchWorkloadVulnerabilities(ctx context.Context, gqlClient graphql.Client, request WorkloadVulnerabilityRequest) (WorkloadVulnerabilityPage, error) {
	page := WorkloadVulnerabilityPage{
		Context: VulnerabilityContext{
			Ingresses:        []VulnerabilityIngress{},
			AuthIntegrations: []string{},
			InboundRules:     []VulnerabilityInboundRule{},
		},
	}
	var image gql.MCPVulnerabilityImage
	var network gql.MCPVulnerabilityNetwork
	switch request.Type {
	case "application":
		response, err := gql.MCPApplicationVulnerabilities(ctx, gqlClient, request.Team, request.Workload, request.Environment, request.First, request.After)
		if err != nil {
			return page, err
		}
		if response.Team.Slug != request.Team {
			return page, fmt.Errorf("scope_mismatch: response does not match the requested team")
		}
		node := response.Team.Environment.Application
		page.Workload = WorkloadIdentity{ID: node.Id, Team: response.Team.Slug, Name: node.Name, Type: request.Type, Environment: node.TeamEnvironment.Environment.Name}
		for _, ingress := range node.Ingresses {
			page.Context.Ingresses = append(page.Context.Ingresses, VulnerabilityIngress{URL: ingress.Url, Type: string(ingress.Type)})
		}
		for _, auth := range node.AuthIntegrations {
			if auth == nil || auth.GetTypename() == nil {
				return WorkloadVulnerabilityPage{}, fmt.Errorf("invalid_response: missing auth integration type")
			}
			page.Context.AuthIntegrations = append(page.Context.AuthIntegrations, *auth.GetTypename())
		}
		image, network = node.Image.MCPVulnerabilityImage, node.NetworkPolicy.MCPVulnerabilityNetwork
	case "job":
		response, err := gql.MCPJobVulnerabilities(ctx, gqlClient, request.Team, request.Workload, request.Environment, request.First, request.After)
		if err != nil {
			return page, err
		}
		if response.Team.Slug != request.Team {
			return page, fmt.Errorf("scope_mismatch: response does not match the requested team")
		}
		node := response.Team.Environment.Job
		page.Workload = WorkloadIdentity{ID: node.Id, Team: response.Team.Slug, Name: node.Name, Type: request.Type, Environment: node.TeamEnvironment.Environment.Name}
		for _, auth := range node.AuthIntegrations {
			if auth == nil || auth.GetTypename() == nil {
				return WorkloadVulnerabilityPage{}, fmt.Errorf("invalid_response: missing auth integration type")
			}
			page.Context.AuthIntegrations = append(page.Context.AuthIntegrations, *auth.GetTypename())
		}
		image, network = node.Image.MCPVulnerabilityImage, node.NetworkPolicy.MCPVulnerabilityNetwork
	default:
		return page, fmt.Errorf("invalid_workload_type: use application or job")
	}
	if page.Workload.Name != request.Workload || page.Workload.Environment != request.Environment || page.Workload.ID == "" {
		return WorkloadVulnerabilityPage{}, fmt.Errorf("scope_mismatch: response does not match the requested workload")
	}
	if image.Id == "" || image.Name == "" || image.Sbom.Status == "" {
		return WorkloadVulnerabilityPage{}, fmt.Errorf("invalid_response: missing image or SBOM metadata")
	}
	for _, rule := range network.Inbound.Rules {
		page.Context.InboundRules = append(page.Context.InboundRules, VulnerabilityInboundRule{Workload: rule.TargetWorkloadName, Team: rule.TargetTeamSlug, Mutual: rule.Mutual})
	}
	page.Snapshot = VulnerabilitySnapshot{
		ImageID: image.Id, ImageName: image.Name, ImageTag: image.Tag, ImageDigest: image.Digest,
		SBOMStatus: string(image.Sbom.Status),
	}
	if summary := image.VulnerabilitySummary; summary != nil {
		page.Snapshot.Total = &summary.Total
		page.Snapshot.SummaryUpdatedAt = summary.LastUpdated
	}
	page.PageInfo = VulnerabilityPageInfo{HasNextPage: image.Vulnerabilities.PageInfo.HasNextPage, EndCursor: image.Vulnerabilities.PageInfo.EndCursor}
	page.Findings = make([]WorkloadVulnerability, 0, len(image.Vulnerabilities.Nodes))
	for _, node := range image.Vulnerabilities.Nodes {
		finding := WorkloadVulnerability{
			ID: node.Id, Identifier: node.Identifier, Package: node.Package, Severity: string(node.Severity),
			Description: node.Description, DetailsLink: node.VulnerabilityDetailsLink,
			RiskAssessment: VulnerabilityRisk{
				CVSSScore: node.RiskAssessment.CvssScore,
				EPSSScore: node.RiskAssessment.EpssScore, EPSSPercentile: node.RiskAssessment.EpssPercentile,
				HasKEVEntry: node.RiskAssessment.HasKevEntry, KnownRansomwareUse: node.RiskAssessment.KnownRansomwareUse,
			},
			Remediation: VulnerabilityRemediation{FixVersion: node.Remediation.FixVersion, LatestVersion: node.Remediation.LatestVersion},
		}
		if node.Suppression != nil {
			state := string(node.Suppression.State)
			finding.SuppressionState = &state
		}
		page.Findings = append(page.Findings, finding)
	}
	return page, nil
}
