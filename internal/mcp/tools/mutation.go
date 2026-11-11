package tools

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ExecuteMutationInput struct {
	Mutation string         `json:"mutation" jsonschema:"required" jsonschema_description:"Exact allowlisted API mutation name. Currently only applyVulnerabilitySuppression is supported; GraphQL documents are not accepted."`
	Input    map[string]any `json:"input" jsonschema:"required" jsonschema_description:"Mutation proposal. For applyVulnerabilitySuppression: target and expectedScopeFingerprint from vulnerabilitySuppressionPreview. Do not supply confirmed or decision; those come only from human elicitation."`
}

func registerMutationTool(s *server.MCPServer, ctx *toolContext) {
	handler := &suppressionHandler{toolContext: ctx, elicit: s.RequestElicitation}
	s.AddTool(mcp.NewTool("execute_mutation",
		mcp.WithDescription("Execute an explicitly allowlisted Nais API mutation after HUMAN confirmation through MCP elicitation. Currently only applyVulnerabilitySuppression is allowed. First query vulnerabilitySuppressionPreview through execute_graphql and review its shared-image scope. Pass the exact mutation name, not a GraphQL document, and input containing only target and expectedScopeFingerprint. The human supplies justification, evidence or explicit risk acceptance, state and authorization. Never infer consent from investigation or autopilot. Without elicitation support, returns an unsaved proposal. Never automatically retry an indeterminate outcome."),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
		mcp.WithInputSchema[ExecuteMutationInput](),
	), handler.handleExecuteMutation)
}

func (h *suppressionHandler) handleExecuteMutation(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ExecuteMutationInput
	if err := request.BindArguments(&args); err != nil {
		return mcp.NewToolResultError("invalid mutation request; no write attempted"), nil
	}
	if args.Mutation != "applyVulnerabilitySuppression" {
		return mcp.NewToolResultError("mutation is not allowed; no write attempted"), nil
	}
	raw, err := json.Marshal(args.Input)
	if err != nil {
		return mcp.NewToolResultError("invalid mutation input; no write attempted"), nil
	}
	var input ConfirmSuppressionInput
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return mcp.NewToolResultError("invalid suppression proposal; only target and expectedScopeFingerprint are accepted; no write attempted"), nil
	}
	return h.handleConfirm(ctx, request, input)
}
