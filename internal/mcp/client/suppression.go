package client

import (
	"context"
	"fmt"

	"github.com/Khan/genqlient/graphql"
	"github.com/nais/cli/internal/naisapi/gql"
)

type SuppressionTarget struct {
	FindingID    string `json:"findingID" jsonschema:"required"`
	AdvisoryID   string `json:"advisoryID" jsonschema:"required"`
	PackagePurl  string `json:"packagePurl" jsonschema:"required"`
	WorkloadID   string `json:"workloadID" jsonschema:"required"`
	Team         string `json:"team" jsonschema:"required"`
	Environment  string `json:"environment" jsonschema:"required"`
	Workload     string `json:"workload" jsonschema:"required"`
	WorkloadType string `json:"workloadType" jsonschema:"required,enum=application,enum=job"`
	ImageID      string `json:"imageID" jsonschema:"required"`
	ImageName    string `json:"imageName" jsonschema:"required"`
	ImageTag     string `json:"imageTag" jsonschema:"required"`
}

type (
	SuppressionPreview = gql.MCPSuppressionPreviewFields
	SuppressionInput   = gql.ApplyVulnerabilitySuppressionInput
	SuppressionResult  = gql.MCPApplySuppressionApplyVulnerabilitySuppressionApplyVulnerabilitySuppressionPayload
)

// SuppressionClient is separate so read-only and mock clients cannot accidentally write.
type SuppressionClient interface {
	PreviewSuppression(context.Context, SuppressionTarget) (SuppressionPreview, error)
	ApplySuppression(context.Context, SuppressionInput) (SuppressionResult, error)
}

func fetchSuppressionPreview(ctx context.Context, c graphql.Client, target SuppressionTarget) (SuppressionPreview, error) {
	response, err := gql.MCPSuppressionPreview(ctx, c, gql.VulnerabilitySuppressionTargetInput(target))
	if err != nil || response == nil {
		return SuppressionPreview{}, fmt.Errorf("suppression preview failed; no write attempted")
	}
	return response.VulnerabilitySuppressionPreview.MCPSuppressionPreviewFields, nil
}

func applySuppression(ctx context.Context, c graphql.Client, input SuppressionInput) (SuppressionResult, error) {
	response, err := gql.MCPApplySuppression(ctx, c, input)
	if err != nil || response == nil {
		return SuppressionResult{}, fmt.Errorf("suppression result unknown; inspect stored state before retrying")
	}
	return response.ApplyVulnerabilitySuppression, nil
}
