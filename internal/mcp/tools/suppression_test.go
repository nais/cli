package tools

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nais/cli/internal/mcp/client"
	"github.com/nais/cli/internal/naisapi/gql"
)

type suppressionTestClient struct {
	client.Client
	preview       client.SuppressionPreview
	writes        int
	input         client.SuppressionInput
	writeError    error
	previewError  error
	indeterminate bool
}

type denySuppressionRequests struct{}

func (denySuppressionRequests) Allow() bool { return false }

func (c *suppressionTestClient) PreviewSuppression(context.Context, client.SuppressionTarget) (client.SuppressionPreview, error) {
	return c.preview, c.previewError
}

func (c *suppressionTestClient) ApplySuppression(_ context.Context, input client.SuppressionInput) (client.SuppressionResult, error) {
	c.writes++
	c.input = input
	return client.SuppressionResult{Persisted: !c.indeterminate, Indeterminate: c.indeterminate, State: input.Decision.State, ScopeFingerprint: input.ExpectedScopeFingerprint}, c.writeError
}

func suppressionHandlerFixture() (*suppressionHandler, *suppressionTestClient, ConfirmSuppressionInput) {
	target := client.SuppressionTarget{
		FindingID: "VUL_finding", AdvisoryID: "CVE-2026-1234", PackagePurl: "pkg:npm/example@1",
		WorkloadID: "APP_app", Team: "team", Environment: "dev", Workload: "app", WorkloadType: "application",
		ImageID: "IMG_image", ImageName: "registry.example/app", ImageTag: "v1",
	}
	c := &suppressionTestClient{preview: client.SuppressionPreview{
		Target: gql.MCPSuppressionPreviewFieldsTargetVulnerabilitySuppressionTarget(target),
		AffectedWorkloads: []gql.MCPSuppressionPreviewFieldsAffectedWorkloadsVulnerabilitySuppressionWorkload{
			{Team: "team", Environment: "dev", Workload: "app", WorkloadType: "application", ImageTag: "v1"},
			{Team: "other", Environment: "prod", Workload: "worker", WorkloadType: "job", ImageTag: "v2"},
		},
		ScopeFingerprint: strings.Repeat("a", 64), IncludesFutureMatches: true,
		Warning: "Shared image scope across other teams, environments, tags and future matches; observed scan-ready workloads only.",
	}}
	h := &suppressionHandler{toolContext: &toolContext{client: c, rateLimiter: allowAll{}}}
	return h, c, ConfirmSuppressionInput{Target: target, ExpectedScopeFingerprint: c.preview.ScopeFingerprint}
}

func acceptedSuppression() *mcp.ElicitationResult {
	return &mcp.ElicitationResult{ElicitationResponse: mcp.ElicitationResponse{
		Action: mcp.ElicitationResponseActionAccept,
		Content: map[string]any{
			"confirm": true, "justification": "Reviewed parser usage",
			"evidence_reference": "https://example.test/review/1", "risk_accepted": false, "state": "NOT_AFFECTED",
		},
	}}
}

func mutationRequest(t *testing.T, input ConfirmSuppressionInput) mcp.CallToolRequest {
	t.Helper()
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var proposal map[string]any
	if err := json.Unmarshal(raw, &proposal); err != nil {
		t.Fatal(err)
	}
	return mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name: "execute_mutation",
		Arguments: map[string]any{
			"mutation": "applyVulnerabilitySuppression",
			"input":    proposal,
		},
	}}
}

func TestSuppressionRequiresHumanConfirmation(t *testing.T) {
	for _, mode := range []string{"evidence", "risk accepted", "declined", "cancelled", "unsupported", "missing consent", "missing evidence", "blank reason", "risk false positive", "malformed", "write error", "indeterminate"} {
		t.Run(mode, func(t *testing.T) {
			h, c, input := suppressionHandlerFixture()
			h.elicit = func(_ context.Context, request mcp.ElicitationRequest) (*mcp.ElicitationResult, error) {
				for _, expected := range []string{input.Target.AdvisoryID, input.Target.PackagePurl, input.Target.ImageID, "worker", "other", "prod", "future matches"} {
					if !strings.Contains(request.Params.Message, expected) {
						t.Errorf("human warning omitted %q", expected)
					}
				}
				result := acceptedSuppression()
				content := result.Content.(map[string]any)
				switch mode {
				case "risk accepted":
					delete(content, "evidence_reference")
					content["risk_accepted"], content["state"] = true, "IN_TRIAGE"
				case "declined":
					result.Action = mcp.ElicitationResponseActionDecline
				case "cancelled":
					result.Action = mcp.ElicitationResponseActionCancel
				case "unsupported":
					return nil, errors.New("client does not support elicitation")
				case "missing consent":
					delete(content, "confirm")
				case "missing evidence":
					delete(content, "evidence_reference")
				case "blank reason":
					content["justification"] = " "
				case "risk false positive":
					content["risk_accepted"], content["state"] = true, "FALSE_POSITIVE"
				case "malformed":
					result.Content = "yes"
				case "write error":
					c.writeError = errors.New("sensitive upstream message")
				case "indeterminate":
					c.indeterminate = true
				}
				return result, nil
			}
			out, err := h.confirm(t.Context(), mcp.CallToolRequest{}, input)
			write := mode == "evidence" || mode == "risk accepted" || mode == "write error" || mode == "indeterminate"
			if (c.writes == 1) != write || c.writes > 1 {
				t.Fatalf("unexpected writes: %d", c.writes)
			}
			if mode == "evidence" || mode == "risk accepted" {
				if err != nil || out.Status != "persisted" || !c.input.Confirmed || c.input.Decision.Justification != "Reviewed parser usage" {
					t.Fatalf("out=%+v error=%v input=%+v", out, err, c.input)
				}
			} else if out.Status == "persisted" {
				t.Fatal("unconfirmed action reported as saved")
			}
			if mode == "write error" || mode == "indeterminate" {
				if err != nil || out.Status != "indeterminate" || strings.Contains(out.Message, "sensitive upstream") {
					t.Fatalf("out=%+v error=%v", out, err)
				}
			}
			raw, _ := json.Marshal(out)
			if strings.Contains(string(raw), "Reviewed parser usage") || strings.Contains(string(raw), "review/1") {
				t.Fatal("decision text leaked into tool output")
			}
		})
	}
}

func TestSuppressionRejectsChangedOrInvalidScope(t *testing.T) {
	for _, mode := range []string{"fingerprint", "target", "missing workload", "duplicate", "api error", "empty purl", "rate limit", "cancelled context"} {
		t.Run(mode, func(t *testing.T) {
			h, c, input := suppressionHandlerFixture()
			h.elicit = func(context.Context, mcp.ElicitationRequest) (*mcp.ElicitationResult, error) {
				t.Fatal("unsafe proposal reached human confirmation")
				return nil, nil
			}
			ctx := t.Context()
			switch mode {
			case "fingerprint":
				input.ExpectedScopeFingerprint = strings.Repeat("b", 64)
			case "target":
				c.preview.Target.ImageTag = "different"
			case "missing workload":
				c.preview.AffectedWorkloads = c.preview.AffectedWorkloads[1:]
			case "duplicate":
				c.preview.AffectedWorkloads = append(c.preview.AffectedWorkloads, c.preview.AffectedWorkloads[0])
			case "api error":
				c.previewError = errors.New("private upstream error")
			case "empty purl":
				input.Target.PackagePurl = ""
			case "rate limit":
				h.rateLimiter = denySuppressionRequests{}
			case "cancelled context":
				cancelled, cancel := context.WithCancel(ctx)
				cancel()
				ctx = cancelled
			}
			out, err := h.confirm(ctx, mcp.CallToolRequest{}, input)
			if c.writes != 0 || out.Status == "persisted" || mode != "fingerprint" && err == nil {
				t.Fatalf("out=%+v error=%v writes=%d", out, err, c.writes)
			}
		})
	}
}

type suppressionElicitor struct{ calls int }

func (e *suppressionElicitor) Elicit(context.Context, mcp.ElicitationRequest) (*mcp.ElicitationResult, error) {
	e.calls++
	return acceptedSuppression(), nil
}

func TestSuppressionProtocolConfirmation(t *testing.T) {
	for _, tc := range []struct {
		name, protocol string
		human          bool
	}{
		{"legacy unsupported", "2025-11-25", false},
		{"legacy human", "2025-11-25", true},
		{"modern unsupported", mcp.LATEST_PROTOCOL_VERSION, false},
		{"modern human", mcp.LATEST_PROTOCOL_VERSION, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			human := tc.human
			h, backend, input := suppressionHandlerFixture()
			s := server.NewMCPServer("test", "1", server.WithToolCapabilities(false))
			registerMutationTool(s, h.toolContext)
			elicitor := &suppressionElicitor{}
			var options []mcpclient.ClientOption
			if human {
				options = append(options, mcpclient.WithElicitationHandler(elicitor))
			}
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			serverRead, clientWrite := io.Pipe()
			clientRead, serverWrite := io.Pipe()
			done := make(chan struct{})
			go func() {
				defer close(done)
				_ = server.NewStdioServer(s).Listen(ctx, serverRead, serverWrite)
			}()
			t.Cleanup(func() {
				cancel()
				_ = clientWrite.Close()
				_ = serverWrite.Close()
				_ = serverRead.Close()
				_ = clientRead.Close()
				<-done
			})
			c := mcpclient.NewClient(transport.NewIO(clientRead, clientWrite, nil), options...)
			if err := c.Start(ctx); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = c.Close() }()
			_, err := c.Initialize(ctx, mcp.InitializeRequest{Params: mcp.InitializeParams{
				ProtocolVersion: tc.protocol, ClientInfo: mcp.Implementation{Name: "test", Version: "1"},
			}})
			if err != nil {
				t.Fatal(err)
			}
			result, err := c.CallTool(ctx, mutationRequest(t, input))
			if err != nil || result.IsError {
				t.Fatalf("result=%+v error=%v", result, err)
			}
			raw, _ := json.Marshal(result.StructuredContent)
			var out SuppressionOutput
			if err := json.Unmarshal(raw, &out); err != nil {
				t.Fatal(err)
			}
			if human {
				if out.Status != "persisted" || backend.writes != 1 || elicitor.calls != 1 {
					t.Fatalf("status=%s message=%s writes=%d human prompts=%d needsInput=%v", out.Status, out.Message, backend.writes, elicitor.calls, result.NeedsInput())
				}
			} else if out.Status != "proposed" || backend.writes != 0 {
				t.Fatalf("model-supplied confirmation authorized a write: %+v", out)
			}
		})
	}
}

func TestSuppressionModernContinuation(t *testing.T) {
	for _, mode := range []string{"success and replay", "expired", "changed input", "forged token", "missing response", "changed scope", "lost capability", "argument consent", "concurrent"} {
		t.Run(mode, func(t *testing.T) {
			h, backend, input := suppressionHandlerFixture()
			ctx := server.WithRequestProtocolInfo(t.Context(), &server.RequestProtocolInfo{
				Modern: true, ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
				ClientCapabilities: &mcp.ClientCapabilities{Elicitation: &mcp.ElicitationCapability{}},
			})
			request := mutationRequest(t, input)
			first, err := h.handleExecuteMutation(ctx, request)
			if err != nil || !first.NeedsInput() || backend.writes != 0 {
				t.Fatalf("first=%+v error=%v writes=%d", first, err, backend.writes)
			}
			token := first.RequestState
			request.Params.MultiRoundTripParams = mcp.MultiRoundTripParams{
				RequestState: token, InputResponses: mcp.InputResponses{"suppression": mcp.NewElicitationInputResponse(*acceptedSuppression())},
			}
			switch mode {
			case "expired":
				pending := h.pending[token]
				pending.Expires = time.Now().Add(-time.Minute)
				h.pending[token] = pending
			case "changed input":
				proposal := request.Params.Arguments.(map[string]any)["input"].(map[string]any)
				proposal["target"].(map[string]any)["findingID"] = "other"
			case "forged token":
				request.Params.RequestState = "forged"
			case "missing response":
				request.Params.InputResponses = nil
			case "changed scope":
				backend.preview.ScopeFingerprint = strings.Repeat("b", 64)
			case "lost capability":
				ctx = server.WithRequestProtocolInfo(t.Context(), &server.RequestProtocolInfo{Modern: true, ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION})
			case "argument consent":
				request.Params.MultiRoundTripParams = mcp.MultiRoundTripParams{}
				args := request.Params.Arguments.(map[string]any)
				args["requestState"], args["inputResponses"] = token, acceptedSuppression()
			}
			if mode == "concurrent" {
				done := make(chan *mcp.CallToolResult, 1)
				go func() { result, _ := h.handleExecuteMutation(ctx, request); done <- result }()
				result, err := h.handleExecuteMutation(ctx, request)
				other := <-done
				if err != nil || result == nil || other == nil || result.IsError == other.IsError || backend.writes != 1 {
					t.Fatalf("concurrent confirmation was not single-use: writes=%d error=%v", backend.writes, err)
				}
				return
			}
			result, err := h.handleExecuteMutation(ctx, request)
			if err != nil {
				t.Fatal(err)
			}
			wantWrites := 0
			if mode == "success and replay" {
				wantWrites = 1
				replay, err := h.handleExecuteMutation(ctx, request)
				if err != nil || !replay.IsError {
					t.Fatalf("continuation reused: %+v error=%v", replay, err)
				}
			}
			if backend.writes != wantWrites {
				t.Fatalf("unsafe continuation writes=%d want=%d result=%+v", backend.writes, wantWrites, result)
			}
		})
	}
}

func TestMutationToolDiscovery(t *testing.T) {
	s := server.NewMCPServer("test", "1", server.WithToolCapabilities(false))
	h, _, _ := suppressionHandlerFixture()
	RegisterTools(s, h.client, allowAll{}, slog.New(slog.NewTextHandler(io.Discard, nil)), true)
	request := []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
	raw, _ := json.Marshal(s.HandleMessage(t.Context(), request))
	var response struct {
		Result struct {
			Tools []struct {
				Name        string         `json:"name"`
				InputSchema map[string]any `json:"inputSchema"`
				Annotations map[string]any `json:"annotations"`
			} `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		t.Fatal(err)
	}
	found := 0
	readOnly := false
	for _, tool := range response.Result.Tools {
		if tool.Name == "execute_graphql" {
			readOnly = tool.Annotations["readOnlyHint"] == true
		}
		if tool.Name != "execute_mutation" {
			continue
		}
		found++
		props := tool.InputSchema["properties"].(map[string]any)
		if len(props) != 2 || props["mutation"] == nil || props["input"] == nil || len(tool.InputSchema["required"].([]any)) != 2 {
			t.Fatalf("unexpected mutation schema: %v", tool.InputSchema)
		}
		if tool.Annotations["readOnlyHint"] != false || tool.Annotations["destructiveHint"] != true || tool.Annotations["idempotentHint"] != false {
			t.Fatalf("unsafe write annotations: %v", tool.Annotations)
		}
	}
	if found != 1 || !readOnly {
		t.Fatalf("found %d mutation tools; read-only query tool: %v", found, readOnly)
	}
}

func TestExecuteMutationRejectsUnsafeInput(t *testing.T) {
	for _, mode := range []string{
		"unknown mutation", "legacy mutation", "graphql document", "multiple mutations",
		"missing mutation", "missing input", "invalid input type", "confirmed", "decision",
		"unknown field", "unknown target field", "missing fingerprint", "invalid target",
	} {
		t.Run(mode, func(t *testing.T) {
			h, backend, input := suppressionHandlerFixture()
			h.elicit = func(context.Context, mcp.ElicitationRequest) (*mcp.ElicitationResult, error) {
				t.Fatal("invalid mutation reached human confirmation")
				return nil, nil
			}
			request := mutationRequest(t, input)
			args := request.Params.Arguments.(map[string]any)
			proposal := args["input"].(map[string]any)
			switch mode {
			case "unknown mutation":
				args["mutation"] = "restartApplication"
			case "legacy mutation":
				args["mutation"] = "updateImageVulnerability"
			case "graphql document":
				args["mutation"] = "mutation { applyVulnerabilitySuppression { persisted } }"
			case "multiple mutations":
				args["mutation"] = []string{"applyVulnerabilitySuppression", "restartApplication"}
			case "missing mutation":
				delete(args, "mutation")
			case "missing input":
				delete(args, "input")
			case "invalid input type":
				args["input"] = "confirmed"
			case "confirmed":
				proposal["confirmed"] = true
			case "decision":
				proposal["decision"] = map[string]any{"state": "NOT_AFFECTED", "justification": "model text"}
			case "unknown field":
				proposal["extra"] = true
			case "unknown target field":
				proposal["target"].(map[string]any)["extra"] = true
			case "missing fingerprint":
				delete(proposal, "expectedScopeFingerprint")
			case "invalid target":
				proposal["target"].(map[string]any)["findingID"] = ""
			}
			result, err := h.handleExecuteMutation(t.Context(), request)
			if err != nil || result == nil || !result.IsError || backend.writes != 0 {
				t.Fatalf("result=%+v error=%v writes=%d", result, err, backend.writes)
			}
		})
	}
}
