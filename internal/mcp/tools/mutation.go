package tools

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ExecuteMutationInput struct {
	Mutation string         `json:"mutation" jsonschema:"required" jsonschema_description:"Exact mutation name: updateImageVulnerability. Not a GraphQL document."`
	Input    map[string]any `json:"input" jsonschema:"required" jsonschema_description:"Suppression proposal containing only vulnerabilityID from the API. The human supplies the custom reason, state and confirmation through elicitation."`
}

func registerMutationTool(s *server.MCPServer, ctx *toolContext) {
	handler := &suppressionHandler{toolContext: ctx, elicit: s.RequestElicitation}
	s.AddTool(mcp.NewTool("execute_mutation",
		mcp.WithDescription("Execute an allowlisted mutation after human confirmation. Suppression has shared-image scope, not workload-local scope. Never infer consent from investigation or autopilot. Without elicitation support, returns an unsaved proposal. Never automatically retry an indeterminate result."),
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
	if args.Mutation != "updateImageVulnerability" {
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
		return mcp.NewToolResultError("invalid suppression proposal; only vulnerabilityID is accepted; no write attempted"), nil
	}
	return h.handleConfirm(ctx, request, input)
}
