package tools

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nais/cli/internal/mcp/client"
	"github.com/nais/cli/internal/naisapi/gql"
)

type PrepareSuppressionInput struct {
	Target client.SuppressionTarget `json:"target" jsonschema:"required"`
}

type ConfirmSuppressionInput struct {
	Target                   client.SuppressionTarget `json:"target" jsonschema:"required"`
	ExpectedScopeFingerprint string                   `json:"expected_scope_fingerprint" jsonschema:"required"`
}

type SuppressionOutput struct {
	Status   string                    `json:"status" jsonschema:"enum=proposed,enum=cancelled,enum=persisted,enum=indeterminate"`
	Proposal client.SuppressionPreview `json:"proposal"`
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
	Expires time.Time
}

func registerSuppressionTools(s *server.MCPServer, ctx *toolContext) {
	handler := &suppressionHandler{toolContext: ctx, elicit: s.RequestElicitation}
	s.AddTool(mcp.NewTool("prepare_vulnerability_suppression",
		mcp.WithDescription("Prepare an UNSAVED suppression for one exact finding and current application/job image. Returns observed affected workloads and the shared image/package/CVE scope across tags, teams, environments and future matches. Requires the backend suppression-preview contract. Does not write or authorize a later write."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithInputSchema[PrepareSuppressionInput](),
		mcp.WithOutputSchema[SuppressionOutput](),
	), mcp.NewStructuredToolHandler(handler.prepare))
	s.AddTool(mcp.NewTool("confirm_vulnerability_suppression",
		mcp.WithDescription("Review the current shared-image suppression proposal and ask the HUMAN through MCP elicitation for justification, evidence or explicit risk acceptance, state and concrete authorization. Never infer consent from investigation or autopilot. No model-supplied confirmation or decision fields. Without elicitation support, returns an unsaved proposal. Writes only the dedicated suppression mutation; never retry an indeterminate outcome automatically."),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
		mcp.WithInputSchema[ConfirmSuppressionInput](),
		mcp.WithOutputSchema[SuppressionOutput](),
	), handler.handleConfirm)
}

func (h *suppressionHandler) handleConfirm(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !mcp.IsModernProtocol(server.RequestProtocolVersion(ctx)) {
		return mcp.NewStructuredToolHandler(h.confirm)(ctx, request)
	}
	var input ConfirmSuppressionInput
	if err := request.BindArguments(&input); err != nil {
		return mcp.NewToolResultError("invalid suppression input; no write attempted"), nil
	}
	attempt := &suppressionHandler{toolContext: h.toolContext}
	info := server.RequestProtocolInfoFromContext(ctx)
	supportsElicitation := info != nil && info.ClientCapabilities != nil && info.ClientCapabilities.Elicitation != nil
	resuming := request.Params.RequestState != ""
	var prompt *mcp.ElicitationRequest
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
	out, err := attempt.confirm(ctx, request, input)
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
	// The opaque, single-use continuation binds client input to the displayed action.
	token := rand.Text()
	h.pending[token] = suppressionPending{Input: input, Expires: now.Add(10 * time.Minute)}
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

func validateSuppressionTarget(target client.SuppressionTarget) error {
	for _, value := range []string{
		target.FindingID, target.AdvisoryID, target.PackagePurl, target.WorkloadID,
		target.Team, target.Environment, target.Workload, target.WorkloadType,
		target.ImageID, target.ImageName,
	} {
		if value == "" || value != strings.TrimSpace(value) || len(value) > 2048 {
			return fmt.Errorf("suppression target must contain exact, nonempty, bounded identifiers")
		}
	}
	if !strings.HasPrefix(target.PackagePurl, "pkg:") ||
		(target.WorkloadType != "application" && target.WorkloadType != "job") ||
		target.ImageTag != strings.TrimSpace(target.ImageTag) || len(target.ImageTag) > 2048 {
		return fmt.Errorf("suppression requires a package purl and application or job")
	}
	return nil
}

func (h *suppressionHandler) preview(ctx context.Context, target client.SuppressionTarget) (SuppressionOutput, error) {
	var out SuppressionOutput
	if !h.rateLimiter.Allow() {
		return out, fmt.Errorf("rate limit exceeded, please try again later")
	}
	if err := ctx.Err(); err != nil {
		return out, err
	}
	if err := validateSuppressionTarget(target); err != nil {
		return out, err
	}
	c, ok := h.client.(client.SuppressionClient)
	if !ok {
		return out, fmt.Errorf("suppression preview is unavailable in this client; no write attempted")
	}
	preview, err := c.PreviewSuppression(ctx, target)
	if err != nil {
		return out, fmt.Errorf("suppression preview failed; no write attempted")
	}
	fingerprint, err := hex.DecodeString(preview.ScopeFingerprint)
	if err != nil || len(fingerprint) != 32 || client.SuppressionTarget(preview.Target) != target ||
		!preview.IncludesFutureMatches || preview.Warning == "" ||
		len(preview.AffectedWorkloads) == 0 || len(preview.AffectedWorkloads) > 50 {
		return out, fmt.Errorf("invalid suppression preview; no write attempted")
	}
	seen := map[string]bool{}
	selected := false
	for _, w := range preview.AffectedWorkloads {
		key := w.Team + "\x00" + w.Environment + "\x00" + w.WorkloadType + "\x00" + w.Workload
		if w.Team == "" || w.Environment == "" || w.Workload == "" || w.ImageTag == "" ||
			(w.WorkloadType != "application" && w.WorkloadType != "job") || seen[key] {
			return out, fmt.Errorf("invalid affected workload list; no write attempted")
		}
		seen[key] = true
		selected = selected || w.Team == target.Team && w.Environment == target.Environment &&
			w.Workload == target.Workload && w.WorkloadType == target.WorkloadType
	}
	if !selected {
		return out, fmt.Errorf("selected workload is missing from suppression scope; no write attempted")
	}
	return SuppressionOutput{Status: "proposed", Proposal: preview, Message: "Unsaved proposal. No suppression has been written."}, nil
}

func (h *suppressionHandler) prepare(ctx context.Context, _ mcp.CallToolRequest, input PrepareSuppressionInput) (SuppressionOutput, error) {
	return h.preview(ctx, input.Target)
}

type suppressionConsent struct {
	Confirm           bool                                   `json:"confirm"`
	Justification     string                                 `json:"justification"`
	EvidenceReference string                                 `json:"evidence_reference"`
	RiskAccepted      bool                                   `json:"risk_accepted"`
	State             gql.ImageVulnerabilitySuppressionState `json:"state"`
}

func (h *suppressionHandler) confirm(ctx context.Context, _ mcp.CallToolRequest, input ConfirmSuppressionInput) (SuppressionOutput, error) {
	if fingerprint, err := hex.DecodeString(input.ExpectedScopeFingerprint); err != nil || len(fingerprint) != 32 {
		return SuppressionOutput{}, fmt.Errorf("the reviewed proposal's scope fingerprint is required")
	}
	out, err := h.preview(ctx, input.Target)
	if err != nil {
		return out, err
	}
	if input.ExpectedScopeFingerprint != out.Proposal.ScopeFingerprint {
		out.Message = "Scope changed. Review this new unsaved proposal before requesting confirmation."
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
		Message: "SUPPRESS THIS FINDING? This changes shared suppression state, not just the selected workload. Review every listed workload and the warning about other teams, tags, environments and future matches.\n" +
			string(preview) + "\nProvide your justification and evidence reference. Without evidence, explicitly accept the risk and choose IN_TRIAGE; this does not establish a false positive or non-applicability. Do not enter secrets or personal information. Confirm only if you authorize this exact action.",
		RequestedSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"confirm":            map[string]any{"type": "boolean", "title": "Authorize this suppression and its shared scope"},
				"justification":      map[string]any{"type": "string", "title": "Why suppress this finding?", "minLength": 1, "maxLength": 4096},
				"evidence_reference": map[string]any{"type": "string", "title": "Supporting evidence reference, if available", "maxLength": 2048},
				"risk_accepted":      map[string]any{"type": "boolean", "title": "I explicitly accept the risk (requires IN_TRIAGE)"},
				"state":              map[string]any{"type": "string", "title": "Suppression state", "enum": []string{"IN_TRIAGE", "FALSE_POSITIVE", "NOT_AFFECTED"}},
			},
			"required": []string{"confirm", "justification", "risk_accepted", "state"},
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
	if strings.TrimSpace(consent.Justification) == "" || len(consent.Justification) > 4096 ||
		len(consent.EvidenceReference) > 2048 ||
		consent.State != gql.ImageVulnerabilitySuppressionStateInTriage &&
			consent.State != gql.ImageVulnerabilitySuppressionStateFalsePositive &&
			consent.State != gql.ImageVulnerabilitySuppressionStateNotAffected ||
		consent.RiskAccepted && consent.State != gql.ImageVulnerabilitySuppressionStateInTriage ||
		!consent.RiskAccepted && strings.TrimSpace(consent.EvidenceReference) == "" {
		return out, fmt.Errorf("justification and evidence or explicit IN_TRIAGE risk acceptance are required; no write attempted")
	}
	if err := confirmationCtx.Err(); err != nil {
		return out, fmt.Errorf("confirmation expired or cancelled; no write attempted")
	}
	var evidence *string
	if strings.TrimSpace(consent.EvidenceReference) != "" {
		evidence = &consent.EvidenceReference
	}
	// Only the elicitation response can supply consent and decision text.
	applied, err := h.client.(client.SuppressionClient).ApplySuppression(confirmationCtx, client.SuppressionInput{
		Target:                   gql.VulnerabilitySuppressionTargetInput(input.Target),
		ExpectedScopeFingerprint: out.Proposal.ScopeFingerprint, Confirmed: true,
		Decision: gql.VulnerabilitySuppressionDecisionInput{
			Justification: consent.Justification, EvidenceReference: evidence,
			RiskAccepted: consent.RiskAccepted, State: consent.State,
		},
	})
	if err != nil || !applied.Persisted || applied.Indeterminate ||
		applied.State != consent.State || applied.ScopeFingerprint != out.Proposal.ScopeFingerprint {
		out.Status, out.Message = "indeterminate", "Suppression is NOT confirmed saved. A write may have happened. Inspect stored state before any retry; do not retry automatically."
		return out, nil
	}
	out.Status, out.Message = "persisted", "The API confirmed the suppression and retained decision. Shared-image scope applies; summary counts on other tags may lag."
	return out, nil
}
