package tools

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nais/cli/internal/mcp/client"
	"github.com/nais/cli/internal/naisapi/gql"
)

type ConfirmSuppressionInput struct {
	VulnerabilityID string `json:"vulnerabilityID"`
}

type SuppressionOutput struct {
	Status   string                    `json:"status"`
	Proposal client.SuppressionFinding `json:"proposal"`
	Message  string                    `json:"message"`
}

type suppressionHandler struct {
	*toolContext
	elicit  func(context.Context, mcp.ElicitationRequest) (*mcp.ElicitationResult, error)
	mu      sync.Mutex
	pending map[string]suppressionPending
}

type suppressionPending struct {
	Input   ConfirmSuppressionInput
	Finding client.SuppressionFinding
	Expires time.Time
}

func (h *suppressionHandler) handleConfirm(ctx context.Context, request mcp.CallToolRequest, input ConfirmSuppressionInput) (*mcp.CallToolResult, error) {
	if !mcp.IsModernProtocol(server.RequestProtocolVersion(ctx)) {
		out, err := h.confirm(ctx, input, nil)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultJSON(out)
	}
	attempt := &suppressionHandler{toolContext: h.toolContext}
	info := server.RequestProtocolInfoFromContext(ctx)
	supportsElicitation := info != nil && info.ClientCapabilities != nil && info.ClientCapabilities.Elicitation != nil
	resuming := request.Params.RequestState != ""
	var prompt *mcp.ElicitationRequest
	var reviewed *client.SuppressionFinding
	if resuming {
		if !supportsElicitation {
			return mcp.NewToolResultError("human confirmation capability is missing; no write attempted"), nil
		}
		h.mu.Lock()
		pending, ok := h.pending[request.Params.RequestState]
		delete(h.pending, request.Params.RequestState)
		h.mu.Unlock()
		if !ok || pending.Input != input || !time.Now().Before(pending.Expires) {
			return mcp.NewToolResultError("suppression confirmation expired, changed or already consumed; no write attempted"), nil
		}
		reviewed = &pending.Finding
		response := server.ElicitationResponse(request.Params.InputResponses, "suppression")
		if response == nil {
			return mcp.NewToolResultError("human confirmation response is missing; no write attempted"), nil
		}
		attempt.elicit = func(context.Context, mcp.ElicitationRequest) (*mcp.ElicitationResult, error) {
			return response, nil
		}
	} else {
		attempt.elicit = func(_ context.Context, request mcp.ElicitationRequest) (*mcp.ElicitationResult, error) {
			prompt = &request
			return nil, nil
		}
	}
	out, err := attempt.confirm(ctx, input, reviewed)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if prompt == nil || resuming || !supportsElicitation {
		return mcp.NewToolResultJSON(out)
	}
	now := time.Now()
	h.mu.Lock()
	if h.pending == nil {
		h.pending = make(map[string]suppressionPending)
	}
	for token, pending := range h.pending {
		if !now.Before(pending.Expires) {
			delete(h.pending, token)
		}
	}
	if len(h.pending) >= 100 {
		h.mu.Unlock()
		out.Message = "Too many pending confirmations. No write attempted."
		return mcp.NewToolResultJSON(out)
	}
	token := rand.Text()
	h.pending[token] = suppressionPending{Input: input, Finding: out.Proposal, Expires: now.Add(10 * time.Minute)}
	h.mu.Unlock()
	out.Message = "Human confirmation required. This proposal has not been saved."
	result, err := mcp.NewToolResultJSON(out)
	if err != nil {
		return nil, err
	}
	result.MultiRoundTripResult = server.NewInputRequestBuilder(token).
		Elicit("suppression", prompt.Params).ToolResult().MultiRoundTripResult
	result.ResultType = mcp.ResultTypeInputRequired
	return result, nil
}

func (h *suppressionHandler) preview(ctx context.Context, id string) (SuppressionOutput, error) {
	var out SuppressionOutput
	if !h.rateLimiter.Allow() {
		return out, fmt.Errorf("rate limit exceeded, please try again later")
	}
	if err := ctx.Err(); err != nil {
		return out, err
	}
	if id == "" || id != strings.TrimSpace(id) || len(id) > 2048 {
		return out, fmt.Errorf("an exact vulnerabilityID is required; no write attempted")
	}
	c, ok := h.client.(client.SuppressionClient)
	if !ok {
		return out, fmt.Errorf("finding lookup is unavailable in this client; no write attempted")
	}
	finding, err := c.GetSuppressionFinding(ctx, id)
	if err != nil {
		return out, fmt.Errorf("finding lookup failed; no write attempted")
	}
	if finding.ID != id || finding.Identifier == "" || finding.Package == "" {
		return out, fmt.Errorf("invalid finding response; no write attempted")
	}
	return SuppressionOutput{Status: "proposed", Proposal: finding, Message: "Unsaved proposal. Shared-image suppression can affect other workloads and tags."}, nil
}

type suppressionConsent struct {
	Confirm bool                                   `json:"confirm"`
	Reason  string                                 `json:"reason"`
	State   gql.ImageVulnerabilitySuppressionState `json:"state"`
}

func (h *suppressionHandler) confirm(ctx context.Context, input ConfirmSuppressionInput, reviewed *client.SuppressionFinding) (SuppressionOutput, error) {
	out, err := h.preview(ctx, input.VulnerabilityID)
	if err != nil {
		return out, err
	}
	if reviewed != nil && !reflect.DeepEqual(out.Proposal, *reviewed) {
		out.Message = "Finding changed. Review this new unsaved proposal before requesting confirmation."
		return out, nil
	}
	preview, err := json.Marshal(out.Proposal)
	if err != nil {
		return out, fmt.Errorf("cannot display suppression proposal; no write attempted")
	}
	confirmationCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	result, err := h.elicit(confirmationCtx, mcp.ElicitationRequest{Params: mcp.ElicitationParams{
		Mode: "form",
		Message: "Suppress this finding?\n" + string(preview) +
			"\nThis changes shared-image suppression, not just one workload. Other teams, tags, environments and future matches may be affected; the API does not provide the complete scope here. Enter your own reason and choose a state. The API records your authenticated Nais identity and includes the reason in activity history. Do not enter secrets or personal information. Confirm only if you authorize this action.",
		RequestedSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"confirm": map[string]any{"type": "boolean", "title": "Authorize this suppression and its shared scope"},
				"reason":  map[string]any{"type": "string", "title": "Your suppression reason", "minLength": 1, "maxLength": 4096},
				"state":   map[string]any{"type": "string", "title": "Suppression state", "description": "IN_TRIAGE: still assessing. FALSE_POSITIVE: incorrect finding. NOT_AFFECTED: vulnerability does not apply. RESOLVED: finding resolved.", "enum": []string{"IN_TRIAGE", "FALSE_POSITIVE", "NOT_AFFECTED", "RESOLVED"}},
			},
			"required": []string{"confirm", "reason", "state"},
		},
	}})
	if err != nil || result == nil {
		out.Message = "Human confirmation unavailable, failed or timed out. No write attempted."
		return out, nil
	}
	if result.Action != mcp.ElicitationResponseActionAccept {
		out.Status, out.Message = "cancelled", "Suppression was not authorized. No write attempted."
		return out, nil
	}
	content, err := json.Marshal(result.Content)
	if err != nil || len(content) > 16384 {
		return out, fmt.Errorf("invalid human response; no write attempted")
	}
	var consent suppressionConsent
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&consent); err != nil {
		return out, fmt.Errorf("invalid human response; no write attempted")
	}
	if !consent.Confirm {
		out.Status, out.Message = "cancelled", "Suppression was not authorized. No write attempted."
		return out, nil
	}
	if strings.TrimSpace(consent.Reason) == "" || len(consent.Reason) > 4096 ||
		consent.State != gql.ImageVulnerabilitySuppressionStateInTriage &&
			consent.State != gql.ImageVulnerabilitySuppressionStateFalsePositive &&
			consent.State != gql.ImageVulnerabilitySuppressionStateNotAffected &&
			consent.State != gql.ImageVulnerabilitySuppressionStateResolved {
		return out, fmt.Errorf("a custom reason and valid state are required; no write attempted")
	}
	if err := confirmationCtx.Err(); err != nil {
		return out, fmt.Errorf("confirmation expired or cancelled; no write attempted")
	}
	current, err := h.preview(confirmationCtx, input.VulnerabilityID)
	if err != nil {
		return out, err
	}
	if !reflect.DeepEqual(out.Proposal, current.Proposal) {
		current.Message = "Finding changed during confirmation. No write attempted; review it again."
		return current, nil
	}
	applied, err := h.client.(client.SuppressionClient).ApplySuppression(confirmationCtx, client.SuppressionInput{
		VulnerabilityID: input.VulnerabilityID,
		Suppress:        true, Reason: consent.Reason, State: &consent.State,
	})
	if err != nil || applied.Id != input.VulnerabilityID || applied.Suppression == nil ||
		applied.Suppression.State != consent.State || applied.Suppression.Reason != consent.Reason {
		out.Status, out.Message = "indeterminate", "Suppression is NOT confirmed saved. A write may have happened. Inspect stored state before any retry; do not retry automatically."
		return out, nil
	}
	out.Status, out.Message = "persisted", "The API confirmed the suppression and custom reason. Its activity history records the authenticated actor and reason. Shared-image scope applies; summary counts may lag."
	return out, nil
}
