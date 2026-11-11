package client

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/Khan/genqlient/graphql"
	"github.com/nais/cli/internal/naisapi/gql"
)

type SuppressionFinding struct {
	TypeName    string            `json:"__typename"`
	ID          string            `json:"id"`
	Identifier  string            `json:"identifier"`
	ImageName   string            `json:"imageName"`
	Package     string            `json:"package"`
	Suppression *SuppressionState `json:"suppression"`
}

type SuppressionState struct {
	State             gql.ImageVulnerabilitySuppressionState `json:"state"`
	ReasonFingerprint [sha256.Size]byte                      `json:"-"`
}

func (s *SuppressionState) UnmarshalJSON(data []byte) error {
	var value struct {
		State  gql.ImageVulnerabilitySuppressionState `json:"state"`
		Reason *string                                `json:"reason"`
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if value.Reason == nil {
		return fmt.Errorf("suppression reason is missing")
	}
	*s = SuppressionState{State: value.State, ReasonFingerprint: sha256.Sum256([]byte(*value.Reason))}
	return nil
}

type (
	SuppressionInput  = gql.UpdateImageVulnerabilityInput
	SuppressionResult = gql.MCPApplySuppressionUpdateImageVulnerabilityUpdateImageVulnerabilityPayloadVulnerabilityImageVulnerability
)

// SuppressionClient is separate so read-only and mock clients cannot accidentally write.
type SuppressionClient interface {
	GetSuppressionFinding(context.Context, string) (SuppressionFinding, error)
	ApplySuppression(context.Context, SuppressionInput) (SuppressionResult, error)
}

func fetchSuppressionFinding(ctx context.Context, c graphql.Client, id string) (SuppressionFinding, error) {
	response, err := gql.MCPSuppressionFinding(ctx, c, id)
	if err != nil || response == nil {
		return SuppressionFinding{}, fmt.Errorf("finding lookup failed; no write attempted")
	}
	var finding SuppressionFinding
	if err := json.Unmarshal(response.Node, &finding); err != nil || finding.TypeName != "ImageVulnerability" {
		return SuppressionFinding{}, fmt.Errorf("vulnerability not found; no write attempted")
	}
	return finding, nil
}

func applySuppression(ctx context.Context, c graphql.Client, input SuppressionInput) (SuppressionResult, error) {
	response, err := gql.MCPApplySuppression(ctx, c, input)
	if err != nil || response == nil || response.UpdateImageVulnerability.Vulnerability == nil {
		return SuppressionResult{}, fmt.Errorf("suppression result unknown; inspect stored state before retrying")
	}
	return *response.UpdateImageVulnerability.Vulnerability, nil
}
